package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	homework "net-c12/homework/v1"

	"google.golang.org/grpc"
)

var addr, certFn, keyFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "listen address")
	flag.StringVar(&certFn, "cert", "serverCert.pem", "certificate file")
	flag.StringVar(&keyFn, "key", "serverKey.pem", "private key file")
}

func main() {
	flag.Parse()

	server := grpc.NewServer()

	rosie := new(Rosie)

	// Register the Rosie service with the server
	homework.RegisterRobotMaidServer(server, rosie.Service())

	// Start the server
	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		log.Fatalf("failed to load certificate: %v", err)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Printf("Listening for TLS connections on %s ...", addr)

	log.Fatal(
		server.Serve(
			// Wrap the listener in a TLS listener
			tls.NewListener(
				listener,
				&tls.Config{
					Certificates:             []tls.Certificate{cert},
					CurvePreferences:         []tls.CurveID{tls.CurveP256},
					MinVersion:               tls.VersionTLS12,
					PreferServerCipherSuites: true,
					NextProtos:               []string{"h2"},
				},
			),
		),
	)
}
