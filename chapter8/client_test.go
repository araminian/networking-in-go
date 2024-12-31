package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {

	resp, err := http.Get("https://www.time.gov")
	if err != nil {
		t.Fatal(err)
	}
	// Although you don’t read the contents of the response body, you must close it
	defer resp.Body.Close()

	// Round to the nearest second to avoid flakiness
	now := time.Now().Round(time.Second)

	date := resp.Header.Get("Date")
	if date == "" {
		t.Fatal("No Date header found")
	}

	// Parse the date string into a time.Time object
	dt, err := time.Parse(time.RFC1123, date)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("time.gov: %s (skew %s)", dt, now.Sub(dt))
}

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	select {} // Block indefinitely
}

func TestBlockIndefinitely(t *testing.T) {

	t.SkipNow()
	// httptest.NewServer creates a new test server that runs a handler
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))

	/*
		Because the helper function http.Get uses the default HTTP client, this
		GET request won’t time out. Instead, the go test runner will eventually time
		out and halt the test, printing the stack trace.
	*/
	_, _ = http.Get(ts.URL)

	t.Fatal("test did not block")

}

func TestBlockIndefinitelyWithTimeout(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// The `http.DefaultClient` is used to send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}

	_ = resp.Body.Close()

}
