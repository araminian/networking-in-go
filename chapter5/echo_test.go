package main

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestEchoServerUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// start the server
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("binding to udp %s: %v", serverAddr, err)
	}

	// cancel the context when the test is done, which will stop the server
	defer cancel()

	// start the client
	// The net.ListenPacket function creates the connection object for both the client and the server.
	client, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	msg := []byte("ping")
	_, err = client.WriteTo(msg, serverAddr)

	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)

	if err != nil {
		t.Fatal(err)
	}

	if addr.String() != serverAddr.String() {
		t.Fatalf("expected server address %s, got %s", serverAddr, addr)
	}

	if !bytes.Equal(buf[:n], msg) {
		t.Errorf("expected %s, got %s", msg, buf[:n])
	}

	t.Logf("replied to %s: %s", addr, string(buf[:n]))

}

// A single UDP connection object can receive packets from more than one sender.
func TestListenPacketUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("binding to udp %s: %v", serverAddr, err)
	}

	defer cancel()

	client, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	// start an interloper, which will send a message to the client
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	/*
		You then create a new UDP connection meant to interlope on the client and echo server and
		interrupt the client . This message should queue up in the client’s receive buffer.
	*/

	interrupt := []byte("pardon me")
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	_ = interloper.Close()
	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	ping := []byte("ping")
	_, err = client.WriteTo(ping, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// the interloper should have interrupted the client
	if !bytes.Equal(buf[:n], interrupt) {
		t.Errorf("expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	t.Logf("replied to %s: %s", addr, string(buf[:n]))

	// addr should be the interloper
	if addr.String() != interloper.LocalAddr().String() {
		t.Errorf("expected message from %q; actual sender is %q",
			interloper.LocalAddr(), addr)
	}

	// now from the server
	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(buf[:n], ping) {
		t.Errorf("expected %s, got %s", ping, buf[:n])
	}

	if addr.String() != serverAddr.String() {
		t.Errorf("expected message from %q; actual sender is %q",
			serverAddr, addr)
	}

	t.Logf("replied to %s: %s", addr, string(buf[:n]))
}

/*
creates the UDP-based net.Conn and demonstrates how net.Conn
encapsulates the implementation details of UDP to emulate a stream-oriented
network connection.
*/
func TestDialUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("binding to udp %s: %v", serverAddr, err)
	}

	defer cancel()

	// create a new UDP connection
	client, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	/*
		interrupts the client by sending a message to it before the
		echo server sends its reply.
	*/

	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	interrupt := []byte("pardon me")

	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	_ = interloper.Close()
	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	/*
		details the difference between a UDP connection using
		net.Conn and one using net.PacketConn,
	*/

	ping := []byte("ping")
	_, err = client.Write(ping)
	if err != nil {
		t.Fatal(err)
	}

	/*
		The client reads packets only from
		the sender address specified in the net.Dial call, as you would expect using a
		stream-oriented connection object. The client never reads the message sent
		by the interloping connection.
	*/
	buf := make([]byte, 1024)
	n, err = client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(buf[:n], ping) {
		t.Errorf("expected %s, got %s", ping, buf[:n])
	}

	/*
		To make sure, you set an ample deadline
		and attempt to read another message from the client.
	*/
	err = client.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("unexpected packet")
	}

}
