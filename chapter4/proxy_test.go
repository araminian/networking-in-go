package main

import (
	"io"
	"net"
	"sync"
	"testing"
)

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup

	// server listens for a "ping" message and responds with a
	// "pong" message. All other messages are echoed back to the client.

	server, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			// handle the connection
			go func(c net.Conn) {
				defer c.Close()

				for {
					buf := make([]byte, 1024)
					n, err := c.Read(buf)
					if err != nil {
						if err == io.EOF {
							return
						}
						t.Error(err)
					}

					// handle the message
					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = c.Write([]byte("pong"))
					default:
						_, err = c.Write([]byte(msg))
					}

					if err != nil {
						if err == io.EOF {
							return
						}
						t.Error(err)
					}
				}

			}(conn)
		}
	}()

	// proxyServer proxies messages from client connections to the
	// destinationServer. Replies from the destinationServer are proxied
	// back to the clients.

	proxyServer, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {

			// client -> proxyServer -> destinationServer
			/*
				Once a client connection accepts 2, the proxy
				establishes a connection to the destination server 3 and starts proxying
				messages
			*/
			conn, err := proxyServer.Accept()
			t.Logf("proxyServer listening on %s", proxyServer.Addr())

			if err != nil {
				return
			}

			go func(from net.Conn) {
				defer from.Close()

				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}

				defer to.Close()

				err = proxy(from, to)
				if err != nil {
					t.Error(err)
				}
			}(conn)
		}
	}()

	// Client
	conn, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	msg := []struct {
		Message string
		Reply   string
	}{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	for i, m := range msg {
		_, err = conn.Write([]byte(m.Message))
		if err != nil {
			t.Error(err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Error(err)
		}

		actual := string(buf[:n])
		t.Logf("%q -> proxy -> %q", m.Message, actual)
		if actual != m.Reply {
			t.Errorf("%d: expected reply: %q; actual: %q",
				i, m.Reply, actual)
		}

	}

	_ = conn.Close()
	_ = proxyServer.Close()
	_ = server.Close()
	wg.Wait()
}
