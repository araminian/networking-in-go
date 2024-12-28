package main

import (
	"net"
	"syscall"
	"time"
)

// DialTimeout is a custom implementation of the `net.Dialer` struct.
// In any case, we are mocking a timeout error.
// Connection attempt will be timed out after the timeout duration.
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{
		Timeout: timeout,
		Control: func(network, address string, c syscall.RawConn) error {
			// Always return a timeout error
			return &net.DNSError{
				Err:         "connection timed out",
				Name:        address,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
	}

	return d.Dial(network, address)
}

func main() {

	ExamplePinger()

}
