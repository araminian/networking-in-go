package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddleware(t *testing.T) {

	handler := http.TimeoutHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			time.Sleep(time.Minute)
		}),
		time.Second,
		"Timed out while reading response",
	)

	r := httptest.NewRequest(http.MethodGet, "http://test/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if actual := string(body); actual != "Timed out while reading response" {
		t.Logf("unexpected body: %q", actual)
	}
}
