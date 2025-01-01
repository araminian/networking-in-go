package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// drainAndClose is a middleware that drains and closes the request body
func drainAndClose(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		},
	)
}

/*
used a multiplexer to send all requests to a single endpoint.
It introduces a slightly more complex multiplexer that has three
endpoints. This one evaluates the requested resource and routes the request
to the right endpoint.
*/
func TestSimpleMux(t *testing.T) {

	// create a new multiplexer
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	serveMux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hello friend.")
	})

	// it's a subpath of /hello
	serveMux.HandleFunc("/hello/there/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Why, hello there.")
	})

	// drainAndClose is a middleware that drains and closes the request body
	/*
		But there’s one issue with the handlers: none of them drain and close the
		request body. This isn’t a big concern in a test like this, but you should stick
		to best practices, nonetheless. If you don’t do so in a real scenario, you may
		cause increased overhead and potential memory leaks.
	*/
	mux := drainAndClose(serveMux)

	testCases := []struct {
		path     string
		response string
		code     int
	}{
		{"http://test/", "", http.StatusNoContent},
		{"http://test/hello", "Hello friend.", http.StatusOK},
		{"http://test/hello/there/", "Why, hello there.", http.StatusOK},
		{"http://test/hello/there", "<a href=\"/hello/there/\">Moved Permanently</a>.\n\n",
			http.StatusMovedPermanently},
		{"http://test/hello/there/you", "Why, hello there.", http.StatusOK},
		{"http://test/hello/and/goodbye", "", http.StatusNoContent},
		{"http://test/something/else/entirely", "", http.StatusNoContent},
		{"http://test/hello/you", "", http.StatusNoContent},
	}

	for i, c := range testCases {
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		resp := w.Result()
		if actual := resp.StatusCode; c.code != actual {
			t.Errorf("%d: expected code %d; actual %d", i, c.code, actual)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if actual := string(b); c.response != actual {
			t.Errorf("%d: expected response %q; actual %q", i,
				c.response, actual)
		}
	}
}
