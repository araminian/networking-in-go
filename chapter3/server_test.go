package main

import (
	"context"
	"io"
	"net"
	"syscall"
	"testing"
	"time"
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

// TestDialTimeout tests the DialTimeout function ,
func TestDialTimeout(t *testing.T) {
	conn, err := DialTimeout("tcp4", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}

	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}

	if !nErr.Timeout() {
		t.Fatal("Error is not a timeout, it is a", nErr.Error())
	}

}

func TestDialContext(t *testing.T) {
	// define a deadline 5 seconds from now
	dl := time.Now().Add(5 * time.Second)

	// create a context with the deadline
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	// cancel the context when the function returns, this will cancel the context and send a cancellation signal to the asynchronous processes
	defer cancel()

	var d net.Dialer

	// Here we are mocking a timeout
	d.Control = func(network, address string, c syscall.RawConn) error {
		// Sleep enough to cause the timeout
		time.Sleep(5 * time.Second)
		return nil
	}

	// DialContext will return a connection or an error
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:http")
	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}

	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("Error is not a timeout, it is a %v", err)
		}
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Error("expected deadline exceeded, got", ctx.Err())
	}
}
