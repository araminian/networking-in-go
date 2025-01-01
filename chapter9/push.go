package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/awoodbeck/gnp/ch09/handlers"
	"github.com/awoodbeck/gnp/ch09/middleware"
)

var (
	addr  = flag.String("listen", "127.0.0.1:8080", "listen address")
	cert  = flag.String("cert", "example.com+5.pem", "certificate")
	pkey  = flag.String("key", "example.com+5-key.pem", "private key")
	files = flag.String("files", "./files", "static file directory")
)

func main() {
	flag.Parse()
	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server gracefully shutdown")
}

func run(addr, files, cert, pkey string) error {
	mux := http.NewServeMux()
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			middleware.RestrictPrefix(
				".", http.FileServer(http.Dir(files)),
			),
		),
	)
	mux.Handle("/",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					/*
						If the http.ResponseWriter is an http.Pusher,
						it can push resources to the client without a
						corresponding request.
					*/
					if pusher, ok := w.(http.Pusher); ok {
						/*
							You specify the path to the resource
							from the client’s perspective, not the file path on the server’s filesystem
							because the server treats the request as if the client originated it to facilitate
							the server push. After you’ve pushed the resources, you serve the response
							for the handler
						*/
						targets := []string{
							"/static/style.css",
							"/static/hiking.svg",
						}
						for _, target := range targets {
							// Push the resources to the client
							if err := pusher.Push(target, nil); err != nil {
								log.Printf("%s push failed: %v", target, err)
							}
						}
					}
					/*
						If, instead, you sent the index.html file before pushing the
						associated resources, the client’s browser may send requests
						for the associated resources before it handles the pushes.
					*/
					http.ServeFile(w, r, filepath.Join(files, "index.html"))
				},
			),
		},
	)
	mux.Handle("/2",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join(files, "index2.html"))
				},
			),
		},
	)

	// start the server
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}
	done := make(chan struct{})
	/*
		When the server receives an os.Interrupt signal, it triggers a call to the
		server’s Shutdown method. Unlike the server’s Close method, which abruptly
		closes the server’s listener and all active connections, Shutdown gracefully shuts
		down the server. It instructs the server to stop listening for incoming connections and blocks until all client connections end. This gives the server the
		opportunity to finish sending responses before stopping the server.
	*/
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for {
			if <-c == os.Interrupt {
				if err := srv.Shutdown(context.Background()); err != nil {
					log.Printf("shutdown: %v", err)
				}
				close(done)
				return
			}
		}
	}()
	log.Printf("Serving files in %q over %s\n", files, srv.Addr)
	var err error
	if cert != "" && pkey != "" {
		log.Println("TLS enabled")
		err = srv.ListenAndServeTLS(cert, pkey)
	} else {
		err = srv.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		err = nil
	}
	<-done
	return err
}
