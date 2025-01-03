package main

import (
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

/*
Shows just a few uses of middleware, such as enforcing
which methods the handler allows, adding headers to the response, or
performing ancillary functions like logging.
*/
func Middleware(next http.Handler) http.Handler {

	/*
		it defines a function that accepts an http.ResponseWriter and
		a pointer to an http.Request and wraps it with an http.HandlerFunc
	*/
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed",
					http.StatusMethodNotAllowed)
			}
			w.Header().Set("X-Content-Type-Options", "nosniff")

			/*
				Likewise, you may want to use middleware to collect metrics,
				ensure specific headers are set on the response,
				or write to a log file before the next handler is called.
			*/
			start := time.Now()
			/*
				In most cases, middleware calls the given handler. But in some cases
				that may not be proper, and the middleware should block the next handler
				and respond to the client itself
			*/
			next.ServeHTTP(w, r)
			log.Printf("Next handler duration %v", time.Now().Sub(start))
		},
	)
}

func RestrictPrefix(prefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// path.Clean removes trailing slashes and removes . and .. elements
			// so we can check if the path is a prefix of the prefix
			for _, p := range strings.Split(path.Clean(r.URL.Path), "/") {
				// if the path is a prefix of the prefix, return a 404
				if strings.HasPrefix(p, prefix) {
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
			}
			next.ServeHTTP(w, r)
		},
	)
}
