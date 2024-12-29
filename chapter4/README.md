## net.Conn interface

You can write powerful network code using the net.Conn interface without having to assert its underlying type, ensuring your code is compatible across operating systems and allowing you to write more robust tests.

The two most useful net.Conn methods are Read and Write. These methods implement the `io.Reader` and `io.Writer` interfaces.

use `net.Conn’s Close` method to close the network connection. This method will return nil if the connection successfully closed or an error otherwise.

The `SetReadDeadline` and `SetWriteDeadline` methods, which accept a `time.Time` object, set the absolute time after which reads and writes on the network connection will return an error. The `SetDeadline` method sets both the read and write deadlines at the same time.

Deadlines allow you to control how long a network connection may remain idle and allow for timely detection of network connectivity problems.

## Reading Data into a Fixed-Size Buffer

To read data from a network connection, you need to provide a buffer for the network connection’s Read method to fill.

```go
buf := make([]byte, 1024)
n, err := conn.Read(buf)
```

The Read method will populate the buffer to its capacity if there is enough data in the connection’s receive buffer. If there are fewer bytes in the receive buffer than the capacity of the buffer you provide, Read will populate the given buffer with the data and return instead of waiting for more data to arrive. In other words, Read is *not* guaranteed to fill your buffer to capacity before it returns.

Check `TestReadIntoBuffer` in `read_test.go` for an example.

## Delimited Reading using a Scanner

Reading data from a network connection by using the method I just showed means your code needs to make sense of the data it receives. Since TCP is a *stream-oriented* protocol, a client can receive a stream of bytes across many packets. Unlike sentences, binary data *doesn’t include inherent punctuation* that tells you **where one message starts and stops**.

If, for example, your code is reading a series of email messages from a server, your code will have to inspect each byte for delimiters indicating the boundaries of each message in the stream of bytes. Alternatively, your client may have an established protocol with the server whereby the server sends a fixed number of bytes to indicate the payload size the server will send next. Your code can then use this size to create an appropriate buffer for the payload.

However, if you choose to use a delimiter to indicate the end of one message and the beginning of another, writing code to handle edge cases isn’t so simple. For example, you may read 1KB of data from a single Read on the network connection and find that it contains two delimiters. This indicates that you have two complete messages, but you don’t have enough information about the chunk of data following the second delimiter to know whether it is also a complete message. If you read another 1KB of data from the network connection and find no delimiters, you can conclude that this entire block of data is a continuation of the last message in the previous 1KB you read. But what if you read 1KB of nothing but delimiters?

If this is starting to sound a bit complex, it’s because you must account for data across multiple Read calls and handle any errors along the way.

In this case, `bufio.Scanner` does what you need.
The `bufio.Scanner` is a convenient bit of code in Go’s standard library that allows you to read delimited data. The Scanner accepts an `io.Reader` as its input. Since `net.Conn` has a `Read` method that implements the `io.Reader` interface, you can use the Scanner to easily read delimited data from a network connection.

Check `TestScanner` in `read_test.go` for an example.

By default, the `bufio.Scanner` will split data read from the network connection when it encounters a newline character (\n) in the stream of data. But you can change this behavior by providing a custom split function like `bufio.ScanWords` which splits the data by whitespace.

Every call to `Scan` can result in multiple calls to the network connection’s `Read` method until the scanner finds its delimiter or reads an error from the connection.

The call to the scanner’s `Text` method returns the chunk of data as a string—a single word and adjacent punctuation, in this case—that it just read from the network connection.

The code continues to iterate around the for loop until the scanner receives an `io.EOF` or other error from the network connection.

## Dynamically Allocating Buffer Size

You can read data of variable length from a network connection, provided that both the sender and receiver have agreed on a protocol for doing so.

The *type-length-value (TLV)* encoding scheme is a good option. TLV encoding uses a *fixed number of bytes* to represent the *type of data*, a *fixed number of bytes* to represent the *value size*, and a *variable number of bytes* to represent the *value itself*.

Our implementation uses a *5-byte header*: 1 byte for the *type* and 4 bytes for the *length*.

The TLV encoding scheme allows you to send a type as a series of bytes to a remote node and constitute the same type on the remote node from the series of bytes.

`binary.Write` is a method that writes the binary representation of the type to the writer. It has the following signature:

```go
func Write(w io.Writer, order binary.ByteOrder, data any) error
```

The `order` parameter is either `binary.BigEndian` or `binary.LittleEndian`, and the `data` parameter is the value to be written.

`binary.Read` is a method that reads the binary representation of the type from the reader. It has the following signature:

```go
func Read(r io.Reader, order binary.ByteOrder, data any) error
```

The `order` parameter is either `binary.BigEndian` or `binary.LittleEndian`, and the `data` parameter is the value to be read.

check the `tlv.go` and `TestPayloads` in the `read_test.go` file for an example.


## Handling Errors While Reading and Writing Data

Unlike writing to file objects, writing to network connections can be unreliable, especially if your network connection is spotty.

Not all errors returned when reading from or writing to a network connection are permanent. The connection can recover from some errors. For example, writing data to a network connection where adverse network conditions delay the receiver’s ACK packets, and where your connection times out while waiting to receive them, can result in a temporary error.

how to check for temporary errors while writing data to a network connection.


Since you might receive a transient error when writing to a network connection, you might need to retry a write operation. 

If the `net.Error’s Temporary` method returns true, the code makes another write attempt by iterating around the for loop. If the error is permanent, the code returns the error.

```go
var (
		err error
		n   int
		i   = 7 // maximum number of retries
	)

	for ; i > 0; i-- {
		n, err = conn.Write([]byte("hello world"))
		if err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Temporary() {
				log.Println("temporary error:", nErr)
				time.Sleep(10 * time.Second)
				continue
			}
			return err
		}
		break
	}
	if i == 0 {
		return errors.New("temporary write failure threshold exceeded")
	}
	log.Printf("wrote %d bytes to %s\n", n, conn.RemoteAddr())

```