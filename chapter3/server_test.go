package main

import (
	"context"
	"io"
	"net"
	"sync"
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

func TestDialContextCancelFanOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Server
	listener, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		// only accept one connection
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	// Clients
	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		// Decrement the wait group counter when the function returns
		defer wg.Done()

		var d net.Dialer

		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		c.Close()
		select {
		// If the context is canceled, we don't need to send anything to the response channel
		case <-ctx.Done():
		// send a message to the response channel
		case response <- id:
		}
	}

	res := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i, &wg)
	}

	response := <-res // it's a blocking call , it will wait for a response from the dialer
	cancel()          // here we cancel the context, which will cancel all the dialers
	wg.Wait()         // wait for all the dialers to finish
	close(res)        // close the response channel

	if ctx.Err() != context.Canceled {
		t.Error("expected canceled context, got", ctx.Err())
	}

	t.Logf("dialer %d retrieved the resource", response)
}

func TestDeadline(t *testing.T) {
	sync := make(chan struct{})

	// server
	listener, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer func() {
			conn.Close()
			close(sync) // read from the sync shouldn't block due to early return
		}()

		// set a read/write deadline
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf) // block until remote nodes sends data

		// first read should timeout, since the deadline is set to 5 seconds and client is not sending any data
		nErr, ok := err.(net.Error)
		if !ok && !nErr.Timeout() {
			t.Errorf("expected timeout, got %v", err)
		}

		// here we signal the client to send data
		sync <- struct{}{}

		// here we forward the deadline to 5 seconds from now
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp4", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	<-sync // here client is waiting for the server to signal it to send data
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1)
	_, err = conn.Read(buf) // block until remote nodes sends data
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}

}
