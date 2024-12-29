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

## Creating Robust Network Applications by Using the io.Package

We learn how to use the `io.Copy`, `io.MultiWriter`, and `io.TeeReader` functions to proxy data between connections, log network traffic, and ping hosts when firewalls attempt to keep you from doing so.

### Proxy
One of the most useful functions from the io package, the `io.Copy` function reads data from an `io.Reader` and writes it to an `io.Writer`. This is useful for creating a proxy, which, in this context, is an intermediary that transfers data between two nodes. 

Check `proxy.go`.

`io.Copy` It writes, to the writer, everything it reads from the reader until the reader returns an `io.EOF`, or, alternately, either the reader or writer `returns an error`. 

The `io.Copy` function returns an error only if a non-`io.EOF` error occurred during the copy, because `io.EOF` means it has read all the data from the reader.

**NOTE**: Since Go version 1.11, if you use `io.Copy` or `io.CopyN` when the source and destination are both `net.TCPConn` objects, the data never enters the user space on Linux, thereby causing the data transfer to occur more efficiently.

### Monitoring a Network Connection

The io package includes useful tools that allow you to do more with network data than just send and receive it using connection objects. For example, you could use `io.MultiWriter` to write a single payload to multiple network connections. You could also use `io.TeeReader` to log data read from a network connection. 

`io.TeeReader` and `io.MultiWriter` expect an `io.Writer`.

`io.TeeReader` returns a new reader that reads from `r` and writes to `w` as well.

`io.MultiWriter` returns a writer that duplicates its writes to all the provided writers.

Although using the `io.TeeReader` and the `io.MultiWriter` in this fashion is powerful, it isn’t without a few caveats. First, both the `io.TeeReader` and the `io.MultiWriter` will block while writing to your writer. Your writer will add latency to the network connection, so be mindful not to block too long. Second, an error returned by your writer will cause the `io.TeeReader` or `io.MultiWriter` to return an error as well, halting the flow of network data.

If you don’t want your use of these objects to potentially interrupt network data flow, I strongly recommend you implement a reader that always returns a nil error and logs its underlying error in a manner that’s actionable.For example, you can modify Monitor’s Write method to always return a
nil error:

```go
func (m *Monitor) Write(p []byte) (int, error) {
	err := m.Output(2, string(p))
	if err != nil {
		log.Println(err) // use the log package’s default Logger
	}
	return len(p), nil
}
```

check `monitor.go`.

## Pinging a Host in ICMP-Filtered Environments

One of its most common uses is to determine whether a host is online by issuing a ping request and receiving a pong reply from the host.

Unfortunately, many internet hosts filter or block ICMP echo replies. If a host filters pongs, the ping erroneously reports that the remote system is unavailable. One technique you can use instead is to establish a TCP connection with the remote host. If you know that the host listens for incoming TCP connections on a specific port, you can use this knowledge to confirm that the host is available, because you can establish a TCP connection only if the host is up and completes the handshake process.

check `port.go`.

## Exploring Go’s TCPConn Object

Accessing the underlying net.TCPConn object allows fine-grained control over the TCP network connection should you need to do such things as modify the read and write buffers, enable keepalive messages, or change the behavior of pending data upon closing the connection. 

```go
tcpConn, ok := conn.(*net.TCPConn)
```

On the server side, you can use the AcceptTCP method on a `net.TCPListener` to accept a connection as a `net.TCPConn` object.

```go
// Retrieving net.TCPConn from the listener
addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:")
if err != nil {
    return err
}

listener, err := net.ListenTCP("tcp", addr)
if err != nil {
    return err
}
tcpConn, err := listener.AcceptTCP()
```

On the client side, use the `net.DialTCP` function:

```go
addr, err := net.ResolveTCPAddr("tcp", "www.google.com:http")
if err != nil {
    return err
}
tcpConn, err := net.DialTCP("tcp", nil, addr)
```

Some of these methods may not be available on your target operating system or may have hard limits imposed by the operating system. My advice is to use the following methods only when necessary. Altering these settings on the connection object from the operating system defaults may lead to network behavior that’s difficult to debug. 

### Controlling Keepalive Messages

A keepalive is a message sent over a network connection to check the connection’s integrity by prompting an acknowledgment of the message from the receiver. After an operating system–specified number of unacknowledged keepalive messages, the operating system will close the connection.

The operating system configuration dictates whether a connection uses keepalives for TCP sessions by default. If you need to enable keepalives on a `net.TCPConn` object, pass true to its `SetKeepAlive` method:

```go
tcpConn.SetKeepAlive(true)
```

You also have control over how often the connection sends keepalive messages using the `SetKeepAlivePeriod` method. This method accepts a `time.Duration` that dictates the keepalive message interval:

```go
tcpConn.SetKeepAlivePeriod(3 * time.Minute)
```

**NOTE**:Using deadlines advanced by a heartbeat is usually the better method
for detecting network problems. As mentioned earlier in this chapter, deadlines provide better cross-platform support, traverse firewalls better, and make sure your application is actively managing the network connection.

### Handling Pending Data on Close

By default, if you’ve written data to `net.Conn` but the data has yet to be sent to or acknowledged by the receiver and you close the network connection, our operating system will complete the delivery in the background. If you don’t want this behavior, the `net.TCPConn` object’s `SetLinger` method allows you to tweak it:

```go
err := tcpConn.SetLinger(-1) // anything < 0 uses the default behavior
```

The `SetLinger` method accepts a `time.Duration` that dictates how long the operating system will wait for the data to be sent or acknowledged by the receiver. If you pass 0, the operating system will discard any pending data and close the connection immediately. If you pass a positive duration, the operating system will wait for the duration to elapse before closing the connection.

If you wish to abruptly discard all unsent data and ignore acknowledgments of sent data upon closing the network connection, set the connection’s linger to zero:

```go
err := tcpConn.SetLinger(0) // immediately discard unsent data on close
```

Setting linger to zero will cause your connection to send an RST packet when your code calls your connection’s Close method, aborting the connection and bypassing the normal teardown procedures.


If you’re looking for a happy medium and your operating system supports it, you can pass a positive integer n to SetLinger. Your operating system will attempt to complete delivery of all outstanding data up to n seconds, after which point your operating system will discard any unsent or unacknowledged data.

```go
err := tcpConn.SetLinger(10) // attempt to deliver data up to 10 seconds
```

### Overriding Default Receive and Send Buffers

Your operating system assigns read and write buffers to each network connection you create in your code. For most cases, those values should be enough. But in the event you want greater control over the read or write buffer sizes, you can tweak their value.

```go
if err := tcpConn.SetReadBuffer(212992); err != nil {
  return err
}
if err := tcpConn.SetWriteBuffer(212992); err != nil {
  return err
}
```

The `SetReadBuffer` method accepts an integer representing the connection’s read buffer size in bytes. Likewise, the `SetWriteBuffer` method accepts an integer and sets the write buffer size in bytes on the connection. Keep in mind that you can’t exceed your operating system’s maximum value for either buffer size.


## Solving Common Go TCP Network Problems

### Zero Window Errors

TCP’s sliding window and how the window size tells the sender how much data the receiver can accept before the next acknowledgment.
A common workflow when reading from a network connection is to read some data from the connection, handle the data, read more data from the connection, handle it, and so on.

But what happens if you don’t read data from a network connection quickly enough? Eventually, the sender may fill the receiver’s receive buffer, resulting in a zero-window state. The receiver will not be able to receive data until the application reads data from the buffer. This most often happens when the handling of data read from a network connection blocks and the code never makes its way around to reading from the socket again,

```go
buf := make([]byte, 1024)
for {
    n, err := conn.Read(buf)
    if err != nil {
        return err
    }
    handle(buf[:n]) // BLOCKS!
}
```

Reading data from the network connection frees up receive buffer space. If the code blocks for an appreciable amount of time while handling the received data, the receive buffer may fill up. A full receive buffer isn’t necessarily bad. Zeroing the window is a way to throttle, or slow, the flow of data from the sender by creating backpressure on the sender. But if it’s unintended or prolonged, a zero window may indicate a bug in your code.

### Sockets Stuck in the CLOSE_WAIT State

Server side of a TCP network connection will enter the CLOSE_WAIT state after it receives and acknowledges the FIN packet from the client. If you see TCP sockets on your server that persist in the CLOSE_WAIT state, it’s likely your code is neglecting to properly call the Close method on its network connections.

```go
for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) { // 1 we never call c.Close() before returning!
			buf := make([]byte, 1024)
			for {
				n, err := c.Read(buf)
				if err != nil {
					return // 2
				}
				handle(buf[:n])
			}
		}(conn)
	}
```


The listener handles each connection in its own goroutine. However,the goroutine fails to call the connection’s Close method before fully returning from the goroutine. Even a temporary error will cause the goroutine to return. And because you never close the connection, this will leave the TCP socket in the CLOSE_WAIT state. If the server attempted to send anything other than a FIN packet to the client, the client would respond with an RST packet, abruptly tearing down the connection. The solution is to make sure to defer a call to the connection’s Close method soon after creating the goroutine.

```go
func (s *Server) handle(c net.Conn) {
	defer c.Close()
	// ...
}
```
