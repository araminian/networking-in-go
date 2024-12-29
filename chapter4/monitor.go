package main

import (
	"io"
	"log"
	"net"
	"os"
)

// Monitor embeds a log.Logger meant for logging network traffic.
type Monitor struct {
	*log.Logger
}

func (m *Monitor) Write(p []byte) (n int, err error) {
	return len(p), m.Output(2, string(p))
}

func ExampleMonitor() {
	// Create a new Monitor object with a logger that writes to stdout.
	monitor := &Monitor{Logger: log.New(os.Stdout, "monitor: ", 0)}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		b := make([]byte, 1024)
		// Create a new TeeReader that reads from the connection and writes to the monitor.
		/*
			This results in an io.Reader that will read from the network
			connection and write all input to the monitor before passing along the input
			to the caller.
		*/
		r := io.TeeReader(conn, monitor)
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
		// Create a new MultiWriter that writes to the connection and the monitor.
		/*
			You log server output by creating an io.MultiWriter that writes to the
			network connection and the monitor.
		*/
		w := io.MultiWriter(conn, monitor)
		_, err = w.Write(b[:n]) // echo the message
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
	}()

	// client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}
	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}
	_ = conn.Close()
	<-done
	/*
		When you send the message Test\n, it’s logged to os.Stdout twice:
		once when you read the message from the connection, and again when
		you echo the message back to the client.
	*/
}
