package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnix(t *testing.T) {

	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	// Path to the socket file , unique to the process
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))

	// start the server asynchronously
	rAddr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}

	// change the permission of the socket file to 0600
	err = os.Chmod(socket, os.ModeSocket|0600)
	if err != nil {
		t.Fatal(err)
	}

	// client
	conn, err := net.Dial("unix", socket)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { conn.Close() }()

	msg := []byte("ping")
	for i := 0; i < 3; i++ {
		_, err := conn.Write(msg)

		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := bytes.Repeat(msg, 3)

	if !bytes.Equal(buf[:n], expected) {

		t.Fatalf("expected reply %q; actual reply %q", expected,
			buf[:n])
	}

	t.Logf("reply from %s: %q", rAddr, buf[:n])

	cancel()
	<-ctx.Done()
}

func TestEchoServerUnixDatagram(t *testing.T) {

	dir, err := os.MkdirTemp("", "echo_unixgram")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	sSocket := filepath.Join(dir, fmt.Sprintf("s%d.sock", os.Getpid()))
	// start the server asynchronously
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	err = os.Chmod(sSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	// client
	cSocket := filepath.Join(dir, fmt.Sprintf("c%d.sock", os.Getpid()))
	client, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = client.Close() }()

	err = os.Chmod(cSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("ping")

	for i := 0; i < 3; i++ {
		_, err := client.WriteTo(msg, serverAddr)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		n, addr, err := client.ReadFrom(buf)
		if err != nil {
			t.Fatal(err)
		}

		if addr.String() != serverAddr.String() {
			t.Fatalf("received reply from %q instead of %q",
				addr, serverAddr)
		}

		if !bytes.Equal(msg, buf[:n]) {
			t.Fatalf("expected reply %q; actual reply %q", msg,
				buf[:n])
		}

		t.Logf("reply from %s: %q", addr, buf[:n])
	}

}
