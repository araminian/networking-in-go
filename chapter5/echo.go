package main

import (
	"context"
	"fmt"
	"net"
)

// context: to allow cancellation of the echo server by the caller
// The caller uses the net.Addr interface to address messages to the echo server.
func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	// Listen for UDP packets on the given address
	//  The net.ListenPacket function is analogous to the net.Listen function you used to create a TCP listener
	s, err := net.ListenPacket("udp", addr)

	if err != nil {
		return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
	}

	// start the server in a goroutine
	go func() {

		// Cancel the context when the server shuts down
		go func() {
			<-ctx.Done()
			s.Close()
		}()

		buf := make([]byte, 1024)

		// read
		for {
			// ReadFrom returns the number of bytes read and the sender's address
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}

			// write , echo back the data to the client
			_, err = s.WriteTo(buf[:n], clientAddr) // server to client
			if err != nil {
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}
