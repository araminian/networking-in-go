package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	// Define the types of payloads
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize uint32 = 10 << 20 // 10 MB
)

var ErrMaxPayloadSize = errors.New("payload size exceeds the maximum allowed")

type Payload interface {
	// Stringer is used to print the payload as a string
	fmt.Stringer
	// ReaderFrom is used to read the payload from an io.Reader
	io.ReaderFrom
	// WriterTo is used to write the payload to an io.Writer
	io.WriterTo
	// Bytes returns the payload as a byte slice
	Bytes() []byte
}

// Binary is a payload that is a byte slice
type Binary []byte

func (m Binary) Bytes() []byte {
	return m
}

func (m Binary) String() string {
	return string(m)
}

func (m Binary) WriteTo(w io.Writer) (int64, error) {

	// binary.Write writes the binary representation of the type to the writer
	err := binary.Write(w, binary.BigEndian, BinaryType) // 1 byte type
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4 bytes length
	if err != nil {
		return 0, err
	}

	n += 4

	o, err := w.Write(m) // sending the payload
	if err != nil {
		return 0, err
	}

	n += int64(o)

	return n, nil
}

// ReadFrom reads the payload from an io.Reader
func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8 // 1 byte type
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	if typ != BinaryType {
		return 0, errors.New("invalid type")
	}

	var size uint32 // 4 bytes length
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}

	n += 4

	if size > MaxPayloadSize {
		return 0, ErrMaxPayloadSize
	}

	// allocate the memory for the payload
	*m = make(Binary, size)

	// read the payload
	o, err := r.Read(*m)
	if err != nil {
		return 0, err
	}

	n += int64(o)

	return n, nil
}

type String string

func (m String) Bytes() []byte {
	return []byte(m)
}

func (m String) String() string {
	return string(m)
}

func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType) // 1 byte type
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4 bytes length
	if err != nil {
		return 0, err
	}

	n += 4

	o, err := w.Write([]byte(m))
	if err != nil {
		return 0, err
	}

	n += int64(o)

	return n, nil
}

func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8 // 1 byte type
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	if typ != StringType {
		return 0, errors.New("invalid type")
	}

	var size uint32 // 4 bytes length
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}

	n += 4

	if size > MaxPayloadSize {
		return 0, ErrMaxPayloadSize
	}

	// allocate the memory for the payload
	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return 0, err
	}

	*m = String(buf)

	n += int64(o)

	return n, nil
}

// Reading arbitrary data from the network

func decode(r io.Reader) (Payload, error) {
	var typ uint8
	// read the type
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	/*
		You must first read a byte from the reader to determine the type and
		create a payload variable to hold the decoded type. If the type you read
		from the reader is an expected type constant, you assign the corresponding type to the payload variable.
	*/
	var payload Payload
	switch typ {
	case BinaryType:
		payload = new(Binary) // new returns a pointer to a new zero value of the type
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("invalid type")
	}

	/*
		You now have enough information to finish decoding the binary data from
		the reader into the payload variable by using its ReadFrom method. But you have a
		problem here. You cannot simply pass the reader to the ReadFrom method. You’ve
		already read a byte from it corresponding to the type, yet the ReadFrom method
		expects the first byte it reads to be the type as well. Thankfully, the io package
		has a helpful function you can use: MultiReader.
		but here you use it to concatenate the byte you’ve
		already read with the reader . From the ReadFrom method’s perspective, it will
		read the bytes in the sequence it expects.
	*/
	_, err = payload.ReadFrom(
		io.MultiReader(bytes.NewReader([]byte{typ}), r),
	)
	/*
		Although the use of io.MultiReader shows you how to inject bytes back
		into a reader, it isn’t optimal in this use case. The proper fix is to remove
		each type’s need to read the first byte in its ReadFrom method. Then, the
		ReadFrom method would read only the 4-byte size and the payload, eliminating the need to inject the type byte back into the reader before passing it
		on to ReadFrom. As an exercise, I recommend you refactor the code to eliminate the need for io.MultiReader
	*/
	if err != nil {
		return nil, err
	}

	return payload, nil
}
