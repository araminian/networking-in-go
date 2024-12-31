package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// Maximum supported datagram size
	// TFTP limits datagram packets to 516 bytes or fewer to avoid fragmentation.
	DatagramSize = 516 // 512 bytes of data + 4 bytes of header

	// Block size for data packets
	BlockSize = DatagramSize - 4 // 4 bytes of header
)

type OpCode uint16 // 2 bytes

const (
	OpRRQ  OpCode = iota + 1 // Read request
	_                        // no Write support
	OpDATA                   // Data
	OpACK                    // Acknowledgment
	OpERR                    // Error
)

type ErrorCode uint16 // 2 bytes

const (
	ErrUnknown ErrorCode = iota + 1
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	ErrIllegalOp
	ErrUnknownID
	ErrFileExists
	ErrNoUser
)

type ReadReq struct {
	Filename string
	Mode     string
}

// Although not used by our server, a client would make use of this method.
// MarshalBinary converts the ReadRequest to a binary representation.
func (q *ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}

	// operation code + filename + null terminator + mode + null terminator
	cap := 2 + 2 + len(q.Filename) + 1 + len(mode) + 1

	b := new(bytes.Buffer)
	// Grow the buffer to the maximum size
	b.Grow(cap)

	// Write the operation code
	err := binary.Write(b, binary.BigEndian, OpRRQ)
	if err != nil {
		return nil, err
	}

	// Write the filename
	_, err = b.WriteString(q.Filename)
	if err != nil {
		return nil, err
	}

	// Write the null terminator
	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	// Write the mode
	_, err = b.WriteString(mode)
	if err != nil {
		return nil, err
	}

	// Write the null terminator
	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil

}

// allows the server to unmarshal a read request from a byte slice, typically read from a network connection with a client.
func (q *ReadReq) UnmarshalBinary(p []byte) error {

	r := bytes.NewBuffer(p)

	var op OpCode
	err := binary.Read(r, binary.BigEndian, &op)
	if err != nil {
		return fmt.Errorf("failed to read operation code: %w", err)
	}

	if op != OpRRQ {
		return errors.New("invalid RRQ")
	}

	// Read the filename until the null terminator which is 0
	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// Remove the null terminator from the end of the filename
	q.Filename = strings.TrimRight(q.Filename, "\x00")
	if q.Filename == "" {
		return errors.New("invalid RRQ")
	}

	// Read the mode until the null terminator which is 0
	q.Mode, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// Remove the null terminator from the end of the mode
	q.Mode = strings.TrimRight(q.Mode, "\x00")
	if q.Mode == "" {
		return errors.New("invalid RRQ")
	}

	actualMode := strings.ToLower(q.Mode)
	if actualMode != "octet" {
		return errors.New("only binary transfers supported")
	}

	return nil
}

// Data

/*
the reasoning being that an io.Reader allows greater flexibility about where you
retrieve the payload. You could just as easily use an *os.File object to read
a file from the filesystem as you could use a net.Conn to read the data from
another network connection.
*/
type Data struct {
	Block   uint16 // 2 bytes
	Payload io.Reader
}

/*
The intention is that the server can keep calling this method to get sequential
blocks of data, each with an increasing block number, from the io.Reader
until it exhausts the reader.

Just like the client, the server needs to monitor
the packet size returned by this method. When the packet size is less than
516 bytes, the server knows it received the last packet and should stop calling
MarshalBinary.
*/
func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	// Increment the block number
	d.Block++

	err := binary.Write(b, binary.BigEndian, OpDATA)
	if err != nil {
		return nil, err
	}

	// Write the block number
	err = binary.Write(b, binary.BigEndian, d.Block)
	if err != nil {
		return nil, err
	}

	// write up to the block size
	_, err = io.CopyN(b, d.Payload, BlockSize)

	if err != nil && err != io.EOF {
		return nil, err
	}

	return b.Bytes(), nil

}

func (d *Data) UnmarshalBinary(p []byte) error {
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var op OpCode
	// read the first 2 bytes as the operation code
	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &op)
	if err != nil || op != OpDATA {
		return errors.New("invalid DATA")
	}

	// read the next 2 bytes as the block number
	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("invalid DATA")
	}

	d.Payload = bytes.NewReader(p[4:])

	return nil

}

// Acknowledgment

// Represents the block number that the client has acknowledged
type Ack uint16 // 2 bytes

func (a Ack) MarshalBinary() ([]byte, error) {
	cap := 2 + 2 // 2 bytes for the operation code + 2 bytes for the block number
	b := new(bytes.Buffer)
	b.Grow(cap)

	err := binary.Write(b, binary.BigEndian, OpACK) // write the operation code
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, a) // write the block number
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(p []byte) error {
	var code OpCode

	r := bytes.NewReader(p)

	err := binary.Read(r, binary.BigEndian, &code) // read operation code
	if err != nil {
		return err
	}

	if code != OpACK {
		return errors.New("invalid ACK")
	}

	return binary.Read(r, binary.BigEndian, a) // read block number
}

// Error

type Error struct {
	Error   ErrorCode
	Message string
}

func (e Error) MarshalBinary() ([]byte, error) {

	// operation code + error code + message + 0 byte
	cap := 2 + 2 + len(e.Message) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)
	err := binary.Write(b, binary.BigEndian, OpERR) // write operation code
	if err != nil {
		return nil, err
	}
	err = binary.Write(b, binary.BigEndian, e.Error) // write error code
	if err != nil {
		return nil, err
	}
	_, err = b.WriteString(e.Message) // write message
	if err != nil {
		return nil, err
	}
	err = b.WriteByte(0) // write 0 byte
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil

}

func (e *Error) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)
	var code OpCode

	err := binary.Read(r, binary.BigEndian, &code) // read operation code
	if err != nil {
		return err
	}
	if code != OpERR {
		return errors.New("invalid ERROR")
	}

	err = binary.Read(r, binary.BigEndian, &e.Error) // read error message
	if err != nil {
		return err
	}

	e.Message, err = r.ReadString(0)
	if err != nil {
		return err
	}
	e.Message = strings.TrimRight(e.Message, "\x00")

	return nil
}
