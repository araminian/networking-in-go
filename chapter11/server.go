package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func NewTLSServer(ctx context.Context, address string,
	maxIdle time.Duration, tlsConfig *tls.Config) *Server {
	return &Server{
		ctx:       ctx,
		ready:     make(chan struct{}),
		addr:      address,
		maxIdle:   maxIdle,
		tlsConfig: tlsConfig,
	}
}

type Server struct {
	ctx context.Context
	/*
		a channel to signal when the server is ready for incoming
		connections.
	*/
	ready     chan struct{}
	addr      string
	maxIdle   time.Duration
	tlsConfig *tls.Config
}

func (s *Server) Ready() {
	if s.ready != nil {
		<-s.ready
	}
}

func (s *Server) ListenAndServeTLS(certFn, keyFn string) error {
	if s.addr == "" {
		s.addr = "localhost:443"
	}
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("binding to tcp %s: %w", s.addr, err)
	}

	/*
		a goroutine to close the listener
		when you cancel the context.
	*/
	if s.ctx != nil {
		go func() {
			<-s.ctx.Done()
			_ = l.Close()
		}()
	}
	return s.ServeTLS(l, certFn, keyFn)
}

func (s Server) ServeTLS(l net.Listener, certFn, keyFn string) error {
	/*
		The ServeTLS method first checks the server’s TLS configuration. If it’s
		nil, it adds a default configuration with PreferServerCipherSuites set to true.
		PreferServerCipherSuites is meaningful to the server only, and it makes the
		server use its preferred cipher suite instead of deferring to the client’s
		preference.
	*/
	if s.tlsConfig == nil {
		s.tlsConfig = &tls.Config{CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		}
	}

	/*
		If the server’s TLS configuration does not have at least one certificate,
		or if its GetCertificate method is nil, you create a new tls.Certificate by
		reading in the certificate and private-key files from the filesystem
	*/
	if len(s.tlsConfig.Certificates) == 0 &&
		s.tlsConfig.GetCertificate == nil {
		cert, err := tls.LoadX509KeyPair(certFn, keyFn)
		if err != nil {
			return fmt.Errorf("loading key pair: %v", err)
		}
		s.tlsConfig.Certificates = []tls.Certificate{cert}
	}

	/*
		The tls.NewListener function acts like
		middleware, in that it augments the listener to return TLS-aware
		connection objects from its Accept method.
	*/
	tlsListener := tls.NewListener(l, s.tlsConfig)
	if s.ready != nil {
		close(s.ready)
	}

	for {
		/*
			Accept returns a new connection from the listener.
		*/
		conn, err := tlsListener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %v", err)
		}
		go func() {
			defer func() { _ = conn.Close() }()
			for {

				/*
					The server handles each connection the same way. It first conditionally sets
					the socket deadline to the server’s maximum idle duration, then
					waits for the client to send data. If the server doesn’t read anything from
					the socket before it reaches the deadline, the connection’s Read method
					returns an I/O time-out error, ultimately causing the connection to close

				*/
				if s.maxIdle > 0 {
					err := conn.SetDeadline(time.Now().Add(s.maxIdle))
					if err != nil {
						return
					}
				}
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					return
				}
				_, err = conn.Write(buf[:n])
				if err != nil {
					return
				}
				/*
					If, instead, the server reads data from the connection, it writes that
					same payload back to the client. Control loops back around to reset the
					deadline and then wait for the next payload from the client
				*/
			}
		}()
	}
}
