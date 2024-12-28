package main

import (
	"io"
	"net"
	"testing"
)

func TestListener(t *testing.T) {
	// if we omit the ip address, the listener will bind to all interfaces (unicast and anycast)
	// tcp : both ipv4 and ipv6
	// tcp4 : ipv4
	// tcp6 : ipv6
	// udp : both ipv4 and ipv6
	// udp4 : ipv4
	// udp6 : ipv6
	listener, err := net.Listen("tcp4", "localhost:0") // 0 means the OS will choose a random port
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		listener.Close()
	}()

	t.Logf("bound to %s", listener.Addr())

	// Handles connections
	// Server is a loop that accepts connections and handles them in a separate goroutine
	for {

		// listener.Accept is a blocking call that waits for a connection
		// conn is the interface that implements the net.Conn interface , it's underlying object is a *TCPConn
		// TCPConn is a struct that implements the net.Conn interface
		// TCPConn provides some additional methods that are specific to TCP connections
		conn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		}

		go func() {
			defer conn.Close() // Close the connection when the function returns (Send FIN to the client)
			// Logic to handle the connection
		}()
	}
}

func TestDial(t *testing.T) {

	// Server Listener
	listener, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})

	// Server goroutine
	go func() {

		// Signal the main goroutine that the server's listener is closed
		defer func() {
			done <- struct{}{}
		}()

		for {

			// After closing the listener, the Accept method will return an error
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			go func(net.Conn) {
				defer func() {
					// Close the connection
					conn.Close()
					// Signal the main goroutine that the connection is closed
					done <- struct{}{}
				}()
				// Read up to 1024 bytes from the socket at a time
				buf := make([]byte, 1024)
				for {
					n, err := conn.Read(buf)
					if err != nil {

						// After the client sends FIN, the server will receive an EOF error
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn)
		}

	}()

	// Client
	// Dial returns a net.Conn interface which here is a *TCPConn
	conn, err := net.Dial("tcp4", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
	<-done // Wait to close the connection
	listener.Close()
	<-done // Wait to close the listener
}
