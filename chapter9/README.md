# BUILDING HTTP SERVICES

In Go, an HTTP server relies on several interacting components: handlers, middleware, and a multiplexer. When it includes all these parts, we call this server a web service.

## The Anatomy of a Go HTTP Server

Check the following picture for the request flow.

![request flow](./flow.png)

First, the server’s `multiplexer` (router, in computer-networking parlance) receives the client’s request. The `multiplexer` determines the destination for the request, then passes it along to the object capable of handling it. We call this object a handler. (The `multiplexer` itself is a handler that routes requests to the most appropriate handler.)

Before the `handler` receives the request, the request may pass through one or more functions called `middleware`. `Middleware` changes the handlers’ behavior or performs auxiliary tasks, such as logging, authentication, or access control.

check `TestSimpleHTTPServer` for more details.

`http.TimeoutHandler` is a middleware that returns a `503 Service Unavailable` error if the request takes too long.

`IdleTimeout` is the maximum amount of time the server will wait for a new request when the connection is idle.

`ReadHeaderTimeout` is the maximum amount of time the server will wait for the request headers.

### Clients Don’t Respect Your Time

Manage the various server time-out values, for the simple reason that clients won’t otherwise respect your server’s time. A client can take its sweet time sending a request to your server. Meanwhile, your server uses resources waiting to receive the request in its entirety. Likewise, your server is at the client’s mercy when it sends the response because it can send data only as fast as the client reads it (or can send only as much as there is TCP buffer space available). Avoid letting the client dictate the duration of a request-response life cycle.

`IdleTimeout` : the length of time clients can remain idle between requests. The `IdleTimeout` field dictates how long the server will keep its side of the TCP socket open while waiting for the next client request when the communication uses keepalives.

`ReadHeaderTimeout` : how long the server should wait to read a request header. It determines how long the server will wait to finish reading the request headers. Keep in mind that this duration has no bearing on the time it takes to read the request body.

If you want to enforce a time limit for reading an incoming request across all handlers, you could manage the request deadline by using the `ReadTimeout` field. If the client hasn’t sent the complete request (the headers and body) by the time the `ReadTimeout` duration elapses, the server ends the TCP connection.

Likewise, you could give the client a finite duration in which to send the request and read the response by using the `WriteTimeout` field.

The `ReadTimeout` and `WriteTimeout` values apply to all requests and responses because they dictate the `ReadDeadline` and `WriteDeadline` values of the TCP socket.

These blanket time-out values may be inappropriate for handlers that expect clients to send large files in a request body or handlers that indefinitely stream data to the client.

In these two examples, the request or response may abruptly time out even if everything went ahead as expected. Instead, a good practice is to rely on the `ReadHeaderTimeout` value.

### Adding TLS Support

HTTP traffic is plaintext by default, but web clients and servers can use HTTP over an encrypted TLS connection, a combination known as HTTPS.

Go’s HTTP server enables `HTTP/2` support over TLS connections only, but enabling TLS is a simple matter.

```go
srv := &http.Server{
	// the convention is to serve HTTPS over port 443, or an augmentation of port 443,
	// like 8443.
	Addr:         "127.0.0.1:8443",
	Handler:      mux,
	IdleTimeout:  5 * time.Minute,
	ReadTimeout:  time.Minute,
	WriteTimeout: time.Minute,
}

l, err := net.Listen("tcp", srv.Addr)
if err != nil {
	t.Fatal(err)
}

go func() {
	// Using the server’s ServeTLS method, you instruct the server to use
	// TLS over HTTP. The ServeTLS method requires the path to both a
	// certificate and a corresponding private key.

    // https://github.com/FiloSottile/mkcert/
	err := srv.ServeTLS(l, "example.com+5.pem", "example.com+5-key.pem")
	if err != http.ErrServerClosed {
		t.Error(err)
	}
}()
```
