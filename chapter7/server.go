package main

import (
	"context"
	"net"
)

// generic stream-based echo server
/*
You’ll be
able to use this function with any streaming network type. That means
you can use it to create a TCP connection to a different node
*/
func streamingEchoServer(ctx context.Context, network string, address string) (net.Addr, error) {

	/*
		Listen accepts a string representing a
		stream-based network and a string representing an address and returns
		an address object and an error interface.
		network: such as tcp, unix, or unixpacket.

		If the network type
		is unix or unixpacket, the address must be the path to a nonexistent file.
	*/
	s, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	// server
	go func() {

		// Wait for the context to be canceled, then close the listener
		go func() {
			<-ctx.Done()
			s.Close()
		}()

		// Accept connections
		for {
			conn, err := s.Accept()
			if err != nil {
				return
			}

			// Handle the connection
			go func() {
				defer func() { conn.Close() }()

				for {
					buf := make([]byte, 1024)
					n, err := conn.Read(buf)
					if err != nil {
						return
					}

					// Echo the data back to the client
					if _, err := conn.Write(buf[:n]); err != nil {
						return
					}
				}

			}()
		}
	}()

	return s.Addr(), nil
}
