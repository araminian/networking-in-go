package main

import (
	"html/template"
	"io"
	"net/http"
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
