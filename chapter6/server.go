package main

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"
)

type Server struct {
	Payload []byte        // the payload served for all read requests
	Retries uint8         // the number of times to retry a failed transmission
	Timeout time.Duration // the duration to wait for an acknowledgment
}

func (s *Server) ListenAndServe(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	log.Printf("listening on %s", conn.LocalAddr())

	return s.Serve(conn)
}

// The server’s Serve method accepts a net.PacketConn and uses it to read incoming requests
// Closing the network connection will cause the method to return.
func (s *Server) Serve(conn net.PacketConn) error {

	if conn == nil {
		return errors.New("nil connection")
	}

	if s.Payload == nil {
		return errors.New("Payload is required")
	}

	if s.Retries == 0 {
		s.Retries = 10
	}

	if s.Timeout == 0 {
		s.Timeout = 6 * time.Second
	}

	var rrq ReadReq
	for {
		buf := make([]byte, DatagramSize)

		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}

		err = rrq.UnmarshalBinary(buf)
		if err != nil {
			log.Printf("[%s] bad request: %v", addr, err)
			continue
		}

		/*
			Since your server is read-only, it’s
			interested only in servicing read requests. If the data read from the connection is a read request, the server passes it along to a handler method in a
			goroutine
		*/
		go s.handle(addr.String(), rrq)
	}
}

func (s *Server) handle(clientAddr string, rrq ReadReq) {
	log.Printf("[%s] requested file: %s", clientAddr, rrq.Filename)

	// connect to the client
	conn, err := net.Dial("udp", clientAddr)
	if err != nil {
		log.Printf("[%s] dial: %v", clientAddr, err)
		return
	}
	defer func() { _ = conn.Close() }()

	var (
		ackPkt  Ack
		errPkt  Err
		dataPkt = Data{Payload: bytes.NewReader(s.Payload)}
		buf     = make([]byte, DatagramSize)
	)

NEXTPACKET:
	for n := DatagramSize; n == DatagramSize; {
		data, err := dataPkt.MarshalBinary()
		if err != nil {
			log.Printf("[%s] preparing data packet: %v", clientAddr, err)
			return
		}

	RETRY:
		for i := s.Retries; i > 0; i-- {
			n, err = conn.Write(data) // send the data packet

			if err != nil {
				log.Printf("[%s] write: %v", clientAddr, err)
				return
			}

			// wait for the client's ACK packet
			_ = conn.SetReadDeadline(time.Now().Add(s.Timeout))

			_, err = conn.Read(buf)
			if err != nil {
				if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
					continue RETRY
				}
				log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
				return
			}

			switch {
			case ackPkt.UnmarshalBinary(buf) == nil:
				if uint16(ackPkt) == dataPkt.Block {
					// received ACK; send next data packet
					continue NEXTPACKET
				}
			case errPkt.UnmarshalBinary(buf) == nil:
				log.Printf("[%s] received error: %v",
					clientAddr, errPkt.Message)
				return
			default:
				log.Printf("[%s] bad packet", clientAddr)
			}
		}
		log.Printf("[%s] exhausted retries", clientAddr)
		return
	}
	log.Printf("[%s] sent %d blocks", clientAddr, dataPkt.Block)
}
