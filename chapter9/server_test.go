package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSimpleHTTPServer(t *testing.T) {

	/*
		Requests sent to the server’s handler first pass through middleware
		named http.TimeoutHandler , then to the handler returned by the handlers
		.DefaultHandler function.

		In this very simple example, you specify only a single
		handler for all requests instead of relying on a multiplexer.

	*/
	srv := &http.Server{
		Addr: "127.0.0.1:8081",
		Handler: http.TimeoutHandler(
			DefaultHandler(),
			2*time.Minute,
			"",
		),
		IdleTimeout:       5 * time.Minute,
		ReadHeaderTimeout: time.Minute,
	}

	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := srv.Serve(l)
		// The Serve method returns http.ErrServerClosed when it closes normally.
		if err != http.ErrServerClosed {
			t.Error(err)
		}
	}()

	testCases := []struct {
		method   string
		body     io.Reader
		code     int
		response string
	}{
		{http.MethodGet, nil, http.StatusOK, "Hello, friend!"},
		{http.MethodPost, bytes.NewBufferString("<world>"), http.StatusOK,
			"Hello, &lt;world&gt;!"},
		{http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
	}

	client := new(http.Client)
	path := fmt.Sprintf("http://%s/", srv.Addr)

	for i, c := range testCases {

		// Create a new request for each test case.
		r, err := http.NewRequest(c.method, path, c.body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		// Send the request to the server.
		resp, err := client.Do(r)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		if resp.StatusCode != c.code {
			t.Errorf("%d: unexpected status code: %q", i, resp.Status)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		_ = resp.Body.Close()
		if c.response != string(b) {
			t.Errorf("%d: expected %q; actual %q", i, c.response, b)
		}

	}
	// Close the server.
	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestHandlerWriteHeader(t *testing.T) {

	/*
		Calling the ResponseWriter’s
		Write method causes Go to make an implicit call to the response’s WriteHeader
		method with http.StatusOK for you. Once the response’s status code is set with
		an explicit or implicit call to WriteHeader, you cannot change it.
	*/
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("bad request"))
		w.WriteHeader(http.StatusBadRequest)
	}

	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	w := httptest.NewRecorder()
	handler(w, r)

	// The status code is 200 OK because the handler writes the response body first.
	t.Logf("Response status: %q", w.Result().Status)

	handler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
	}

	w = httptest.NewRecorder()
	handler(w, r)

	// The status code is 400 Bad Request because the handler sets the status code first.
	t.Logf("Response status: %q", w.Result().Status)
}
