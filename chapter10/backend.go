package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("listen", "localhost:8080", "listen address")

/*
The /style.css and /hiking.svg resources do not include a full
URL (such as http://localhost:2020/style.css) because the backend web service
does not know anything about Caddy or how clients access Caddy. When
you exclude the scheme, hostname, and port number in the resource
address, the client’s web browser should encounter /style.css in the HTML
and prepend the scheme, hostname, and port number it used for the initial
request before sending the request to Caddy.
*/
var index = []byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Caddy Backend Test</title>
<link href="/style.css" rel="stylesheet">
</head>
<body>
<p><img src="/hiking.svg" alt="hiking gopher"></p>
</body>
</html>`)

func main() {
	flag.Parse()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	err := run(*addr, c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server stopped")
}

func run(addr string, c chan os.Signal) error {
	mux := http.NewServeMux()
	mux.Handle("/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			/*
				The web service receives all requests from Caddy, no matter which
				client originated the request. Likewise, it sends all responses back to Caddy,
				which then routes the response to the right client. Conveniently, Caddy adds
				an X-Forwarded-For header to each request with the originating client’s IP
				address. Although you don’t do anything other than log this information,
				your backend service could use this IP address to differentiate between
				client requests. The service could deny requests based on client IP address,
				for example.
			*/
			clientAddr := r.Header.Get("X-Forwarded-For")
			log.Printf("%s -> %s -> %s", clientAddr, r.RemoteAddr, r.URL)
			_, _ = w.Write(index)
		}),
	)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}
	go func() {
		for {
			if <-c == os.Interrupt {
				_ = srv.Close()
				return
			}
		}
	}()
	fmt.Printf("Listening on %s ...\n", srv.Addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}
