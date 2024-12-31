package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type User struct {
	First string
	Last  string
}

func handlePostUser(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Close the body and discard the contents
		/*
			Unlike the Go HTTP client, the Go HTTP server must explicitly drain
			the request body before closing it.
		*/
		defer func(r io.ReadCloser) {
			_, _ = io.Copy(io.Discard, r)
			r.Close()
		}(r.Body)

		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		var u User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			t.Error(err)
			http.Error(w, "Decode Failed", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func TestPostUser(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(handlePostUser(t)))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}

	buf := new(bytes.Buffer)

	u := User{First: "John", Last: "Doe"}

	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Expected status %d, got %d", http.StatusAccepted, resp.StatusCode)
	}

	_ = resp.Body.Close()
}
