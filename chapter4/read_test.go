package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"reflect"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	// << means bitwise left shift, 1 << 24 is 2^24
	payload := make([]byte, 1<<24) // 16MB

	// generate a random payload
	_, err := rand.Read(payload)
	if err != nil {
		t.Fatal(err)
	}

	// server
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// server goroutine, accept only one connection and write the payload to it
	go func() {
		conn, err := listener.Accept() // Blocking call , wait for a connection
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()

		// write the payload to the connection
		_, err = conn.Write(payload)
		if err != nil {
			t.Log(err)
		}
	}()

	// client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1<<19) // 512KB

	// read loop, as server sends data in chunks of 16MB, we need to read in chunks
	/*
		The client then reads up to the first 512KB from the
		connection, before continuing around the loop. The client continues to
		read up to 512KB at a time until either an error occurs or the client reads
		the entire 16MB payload.
	*/
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}

		t.Logf("read %d KB", n/1024)
	}

	conn.Close()
	listener.Close()

}

func TestScanner(t *testing.T) {

	const payload = "The bigger the interface, the weaker the abstraction."

	// server
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// server goroutine, accept only one connection and write the payload to it
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()
		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Errorf("failed to write payload to connection: %v", err)
		}
	}()

	// client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// create a scanner that reads from the connection
	scanner := bufio.NewScanner(conn)

	// split the scanner into words which are delimited by whitespace
	scanner.Split(bufio.ScanWords)

	var words []string

	// read the words into the slice, it will block until the connection is closed
	// it returns false when the connection is closed, io.EOF or any other error
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	// check for any errors
	err = scanner.Err()
	if err != nil {
		t.Errorf("failed to read words from connection: %v", err)
	}

	expected := []string{"The", "bigger", "the", "interface,", "the", "weaker", "the", "abstraction."}

	if !reflect.DeepEqual(words, expected) {
		t.Errorf("expected %v, got %v", expected, words)
	}

	t.Logf("Scanned %v", words)
}

/*
illustrates how you can send your two distinct types over a network connection
and properly decode them back into their original type on the receiver’s end.
*/
func TestPayloads(t *testing.T) {
	b1 := Binary("Clear is better than clever.")
	b2 := Binary("Don't panic.")

	s1 := String("Errors are values.")

	payloads := []Payload{&b1, &b2, &s1}

	// server
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// server goroutine, accept only one connection and write the payloads to it
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()

		for _, payload := range payloads {
			_, err := payload.WriteTo(conn)
			if err != nil {
				t.Errorf("failed to write payload to connection: %v", err)
				break
			}
		}
	}()

	// client
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < len(payloads); i++ {
		actual, err := decode(conn)
		if err != nil {
			t.Errorf("failed to decode payload %d: %v", i, err)
		}

		if expected := payloads[i]; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v, got %v", expected, actual)
			continue

		}

		t.Logf("[%T] %[1]q", actual)
	}

}

// TestMaxPayloadSize tests the maximum payload size
func TestMaxPayloadSize(t *testing.T) {
	buf := new(bytes.Buffer)
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 4 bytes indicating the size of the payload is 1 GB
	if err != nil {
		t.Fatal(err)
	}
	var b Binary
	_, err = b.ReadFrom(buf)
	if err != ErrMaxPayloadSize {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
