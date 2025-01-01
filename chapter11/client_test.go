package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

func TestClientTLS(t *testing.T) {
	/*
		the httptest.NewTLSServer function handles the
		HTTPS server’s TLS configuration details, including the creation of a new
		certificate.
	*/
	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.TLS == nil {
					/*
						If the server receives the client’s request over HTTP, the request’s TLS
						field will be nil. You can check for this case  and redirect the client to the
						HTTPS endpoint accordingly
					*/
					u := "https://" + r.Host + r.RequestURI
					http.Redirect(w, r, u, http.StatusMovedPermanently)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer ts.Close()
	/*
		For testing purposes, the server’s Client method returns a new *http
		Client that inherently trusts the server’s certificate. You can use this client
		to test TLS-specific code within your handlers.
	*/
	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}

	// using a client that does not trust the server's certificate

	/*
		You override the default TLS configuration in your client’s transport by
		creating a new transport, defining its TLS configuration, and configuring
		http2 to use this transport. It’s good practice to restrict your client’s curve
		preference to the P-256 curve

	*/
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
		},
	}

	/*
		Since your transport no longer relies on the default TLS configuration,
		the client no longer has inherent HTTP/2 support. You need to explicitly
		bless your transport with HTTP/2 support if you want to use it. Of course,
		this test doesn’t rely on HTTP/2, but this implementation detail can trip
		you up if you’re unaware that overriding the transport’s TLS configuration
		removes HTTP/2 support.
	*/

	err = http2.ConfigureTransport(tp)
	if err != nil {
		t.Fatal(err)
	}

	client2 := &http.Client{Transport: tp}

	_, err = client2.Get(ts.URL)
	if err == nil || !strings.Contains(err.Error(),
		"certificate signed by unknown authority") {
		t.Fatalf("expected unknown authority error; actual: %q", err)
	}

	/*
		The first call
		to the test server results in an error because your client doesn’t trust the
		server certificate’s signatory. You could work around this and configure
		your client’s transport to skip verification of the server’s certificate
		by setting its InsecureSkipVerify field to true

	*/
	tp.TLSClientConfig.InsecureSkipVerify = true

	resp, err = client2.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
}

func TestClientTLSGoogle(t *testing.T) {

	/*
		You use the `tls.DialWithDialer` function to initiate a TLS connection with
		Google’s server. This function returns a `net.Conn` object that you can use
		to retrieve the TLS connection’s state.
	*/
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 30 * time.Second},
		"tcp",
		"www.google.com:443",
		&tls.Config{
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	/*
		The underlying TLS client used the operating system’s
		trusted certificate storage
	*/
	state := conn.ConnectionState()
	t.Logf("TLS 1.%d", state.Version-tls.VersionTLS10)
	t.Log(tls.CipherSuiteName(state.CipherSuite))
	t.Log(state.VerifiedChains[0][0].Issuer.Organization[0])
	_ = conn.Close()
}
