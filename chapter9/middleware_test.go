package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestRestrictPrefix(t *testing.T) {

	// create temp dir for testing
	dir, err := os.MkdirTemp("", "test-restrict-prefix")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testCases := []struct {
		path string
		code int
	}{
		{"http://test/static/sage.svg", http.StatusOK},
		{"http://test/static/.secret", http.StatusNotFound},
		{"http://test/static/.dir/secret", http.StatusNotFound},
	}

	// create test files
	os.WriteFile(filepath.Join(dir, "sage.svg"), []byte("sage"), 0644)
	os.WriteFile(filepath.Join(dir, ".secret"), []byte("secret"), 0644)
	os.Mkdir(filepath.Join(dir, ".dir"), 0755)
	os.WriteFile(filepath.Join(dir, ".dir", "secret"), []byte("secret"), 0644)

	// StripPrefix removes the /static/ prefix from the request path
	handler := http.StripPrefix("/static/", RestrictPrefix(".", http.FileServer(http.Dir(dir))))

	for _, tc := range testCases {
		r := httptest.NewRequest(http.MethodGet, tc.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		resp := w.Result()
		if resp.StatusCode != tc.code {
			t.Errorf("Expected status %d, got %d", tc.code, resp.StatusCode)
		}
	}

}
