package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEchoServerTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverAddress := "localhost:34443"
	maxIdle := time.Second
	server := NewTLSServer(ctx, serverAddress, maxIdle, nil)
	done := make(chan struct{})
	go func() {
		err := server.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil && !strings.Contains(err.Error(),
			"use of closed network connection") {
			t.Error(err)
			return
		}
		done <- struct{}{}
	}()
	// block until it’s ready for incoming connections
	server.Ready()

	/*
		Pinning a server certificate to the client is straightforward. First, you
		read in the cert.pem file. Then, you create a new certificate pool and
		append the certificate to it. Finally, you add the certificate pool to the
		tls.Config’s RootCAs field. As the name suggests, you can add more than
		one trusted certificate to the certificate pool. This can be useful when
		you are migrating to a new certificate but have yet to completely phase
		out the old certificate.

		The client, using this configuration, will authenticate only servers that
		present the cert.pem certificate or any certificate signed by it.
	*/
	cert, err := os.ReadFile("cert.pem")
	if err != nil {
		t.Fatal(err)
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		t.Fatal("failed to append certificate to pool")
	}
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
		RootCAs:          certPool,
	}

	// client

	/*
		You pass tls.Dial the tls.Config with the pinned server certificate.
		Your TLS client authenticates the server’s certificate without having
		to resort to using InsecureSkipVerify and all the insecurity that option
		introduces.
	*/
	conn, err := tls.Dial("tcp", serverAddress, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}
	hello := []byte("hello")
	_, err = conn.Write(hello)
	if err != nil {
		t.Fatal(err)
	}
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	if actual := b[:n]; !bytes.Equal(hello, actual) {
		t.Fatalf("expected %q; actual %q", hello, actual)
	}
	time.Sleep(2 * maxIdle)
	_, err = conn.Read(b)
	if err != io.EOF {
		t.Fatal(err)
	}
	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	<-done
}
