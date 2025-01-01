package main

import (
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strings"
)

/*
This code could have a security vulnerability since part of the response
body might come from the request body. A malicious client can send a
request payload that includes JavaScript, which could run on a client’s
computer. This behavior can lead to an XSS attack.

To prevent these attacks, you must properly escape all client-supplied content
before sending it in a response. Here, you use the html/template package
to create a simple template that reads Hello, {{.}}!, where {{.}} is a
placeholder for part of your response. Templates derived from the html/template
package automatically escape HTML characters when you populate them and
write the results to the response writer
*/
var t = template.Must(template.New("hello").Parse("Hello, {{.}}!"))

func DefaultHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			// drains and closes the request body
			defer func(r io.ReadCloser) {
				_, _ = io.Copy(io.Discard, r)
				_ = r.Close()
			}(r.Body)

			var b []byte

			switch r.Method {
			case http.MethodGet:
				b = []byte("friend")
			case http.MethodPost:
				var err error
				b, err = io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Internal server error",
						http.StatusInternalServerError)
					return
				}
			default:
				// not RFC-compliant due to lack of "Allow" header
				http.Error(w, "Method not allowed",
					http.StatusMethodNotAllowed)
				return
			}
			_ = t.Execute(w, string(b))
		},
	)
}

// The Methods type is a multiplexer (router)
// since it routes requests to the appropriate handler.
type Methods map[string]http.Handler

func (h Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// drains and closes the request body
	defer func(r io.ReadCloser) {
		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}(r.Body)

	if handler, ok := h[r.Method]; ok {
		if handler == nil {
			http.Error(w, "Internal server error",
				http.StatusInternalServerError)
		} else {
			/*
				ServeHTTP method
				to implement the http.Handler interface, so you can use Methods as a handler
				itself.
			*/
			handler.ServeHTTP(w, r)
		}
		return
	}

	// If the request method isn’t in the map, ServeHTTP responds with the Allow
	// header and a list of supported methods in the map. All that’s left do now
	// is determine whether the client explicitly requested the OPTIONS method.
	w.Header().Add("Allow", h.allowedMethods())
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func (h Methods) allowedMethods() string {
	a := make([]string, 0, len(h))
	for k := range h {
		a = append(a, k)
	}
	sort.Strings(a)
	return strings.Join(a, ", ")
}

// DefaultMethodsHandler returns a handler that supports GET and POST methods.
// It uses the Methods type to implement the http.Handler interface.
// A Handler is an interface that has a ServeHTTP method.
func DefaultMethodsHandler() http.Handler {
	return Methods{
		http.MethodGet: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("Hello, friend!"))
			},
		),

		http.MethodPost: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Internal server error",
						http.StatusInternalServerError)
					return
				}
				_, _ = fmt.Fprintf(w, "Hello, %s!",
					html.EscapeString(string(b)))
			},
		),
	}

}
