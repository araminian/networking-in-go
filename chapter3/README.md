## Theory

- `receive buffer`: A receive buffer is block of memory that stores incoming data from a network interface card (NIC). It allows the node to accept certain amount of data without requiring an application to read the data immediately. We have per-connection receive buffer. When in our `Go Code` we call `Read` from `net.Conn` interface, it will read data from the receive buffer.

- `window size` : The window size is the maximum amount of data that can be sent without receiving an acknowledgment. It is a measure of the sender's buffer size and the receiver's buffer size. The window size is used to control the flow of data between the sender and receiver. A window size of `0` means the receiver's buffer is full and the sender should wait for an acknowledgment before sending more data. Each server and client has a window size.

- `RST`: informs tha sender that the receiver's side of the connection closed and will not accept any more data. the sender should close the connection. Intermediate nodes like routers and firewalls can send a RST packet to close a connection.


## Code

### Errors
- `Errors` returned from functions and methods in the `net` package typically implement the `net.Error` interface. Which includes `Temporary() bool` and `Timeout() bool` methods.

  - `Timeout() bool` returns true if resource is temporarily unavailable, and call would block, or the connection timed out.
  - `Temporary() bool` returns true if `Timeout()` is true, the function call was interrupted by a network error that may have been resolved, or there are too many open files (exceeds the system limit).

  ```go
  if nErr, ok := err.(net.Error); ok && nErr.Temporary() {
    // Handle temporary error
  }
  ```

  `net.OpError` is a struct that implements the `net.Error` interface. It contains the `Op` field which is a string that describes the operation that failed, and the `Net` field which is a string that describes the network type.

### Timeout
- the default `Dial` function doesn't have a timeout. It's better to Control the timeout, so we can use `net.DialTimeout` function.

  - `net.Dialer` is a struct that implements the `net.Dialer` interface. It contains the `Timeout` field which is a time.Duration that describes the timeout.
  - `Control` is a function that is called with the network, address, and a `syscall.RawConn` interface. It returns an error. The `syscall.RawConn` interface is a wrapper around a `syscall.Conn` struct. The `syscall.Conn` struct is a wrapper around a `net.Conn` struct.
  - `Control` is called after creating the socket, but before the connection is established.
  - `Control` Use Case:
    - We can read the socket options, or set the socket options.

### Timeout using Context

A better way to handle timeouts is to use the `context` package. By using `context` we can send cancellation signals to the `asynchronous` processes. or we can set a deadline for the operation. We can call `cancel function` even before reaching the `deadline`.

`dialContext` is a function that is similar to `dial` but it takes a `context` as an argument. We can cancel the context to cancel the operation.

`context.Err()` returns the error that caused the context to be canceled. `context.DeadlineExceeded` is a constant that is returned when the context is canceled due to a deadline being exceeded. 

`context.WithDeadline` returns a `context.Context` and a `context.CancelFunc`. We can call the `CancelFunc` to cancel the context.


We can pass the `context` to multiple dialers, and cancel all the calls at the same time by calling the `CancelFunc`. An example is, we need get a file from multiple servers, so we use multiple dialers to get the file from each server. If one of the dialers gets the file, we can cancel the other dialers. Check `TestDialContextCancelFanOut` for more details.