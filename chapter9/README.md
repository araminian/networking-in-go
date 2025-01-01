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

## Handlers

When a client sends a request to an HTTP server, the server needs to figure out what to do with it. The server may need to retrieve various resources or perform an action, depending on what the client requests. A common design pattern is to specify bits of code to handle these requests, known as handlers.

In Go, handlers are objects that implement the `http.Handler` interface. They read client requests and write responses. The `http.Handler` interface consists of a single method to receive both the response and the request:

```go
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}
```

We often define handlers as functions, as in this common pattern:

Here, you wrap a function that accepts an `http.ResponseWriter` and an `http.Request` pointer in the `http.HandlerFunc` type, which implements the `Handler` interface.

```go
handler := http.HandlerFunc(
    func(w http.ResponseWriter, r *http.Request) {
    _, _ = w.Write([]byte("Hello, world!"))
    },
)
```

Go programmers commonly convert a function with the signature `func(w http.ResponseWriter, r *http.Request)` to the `http.HandlerFunc` type so the function implements the `http.Handler` interface.

**NOTE** it’s important for the server to do the same with the request body. But unlike the Go HTTP client, closing the request body does not implicitly drain it. Granted, the `http.Server` will close the request body for you, but it won’t drain it. To make sure you can reuse the `TCP session`, I recommend you `drain the request body` at a minimum. `Closing` it is optional.

### Test Your Handlers with httptest

`net/http/httptest`. This package makes unit-testing handlers painless.

The `net/http/httptest` package exports a NewRequest function that accepts an HTTP method, a target resource, and a request body io.Reader. It returns a pointer to an http.Request ready for use in an http.Handler:

```
func NewRequest(method, target string, body io.Reader) *http.Request
```

Unlike its `http.NewRequest` equivalent, `httptest.NewRequest` will panic instead of returning an error. This is preferable in tests but not in production code.


The `httptest.NewRecorder` function returns a pointer to an `httptest.ResponseRecorder`, which implements the `http.ResponseWriter` interface.

Although the `httptest.ResponseRecorder` exports fields that look tempting to use (I don’t want to tempt you by mentioning them), I recommend you call its `Result` method instead. The `Result` method returns a pointer to an `http.Response` object, just like the one we used in the last chapter. As the method’s name implies, it waits until the handler returns before retrieving the `httptest.ResponseRecorder`‘s results.

If you’re interested in performing integration tests, the `net/http/httptest` package includes a test server implementation.

```go
func TestDefaultHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	DefaultHandler().ServeHTTP(rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
}
```

### How You Write the Response Matters

Here’s one potential pitfall: the order in which you write to the response body and set the response status code matters. The client receives the response status code first, followed by the response body from the server.

If you write the response body first, Go infers that the response status code is 200 OK and sends it along to the client before sending the response body.

check `TestHandlerWriteHeader` for more details.

Remember, the server sends the response status code before the response body. Once the response’s status code is set with an explicit or implicit call to `WriteHeader`, you cannot change it because it’s likely on its way to the client.

This function sets the content type to text/plain, sets the status code to `400 Bad Request`, and writes the error message to the response body.

```go
func badRequestHandler(w http.ResponseWriter, r *http.Request) {
  http.Error(w, "Bad request", http.StatusBadRequest)
}
```

### Any Type Can Be a Handler

Because http.Handler is an interface, you can use it to write powerful constructs for handling client requests.

Let’s improve upon the `default handler`. check `DefaultMethodsHandler` for more details.

### Injecting Dependencies into Handlers

The `http.Handler` interface gives you access to the request and response objects.
But it’s likely you’ll require access to additional functionality like a logger,
metrics, cache, or database to handle a request.

For example, you may want to inject a logger to record request errors or inject a database object to retrieve data used to create the response. The easiest way to inject an object into a handler is by using a closure.


Following demonstrates how to inject a `SQL database object` into an `http.Handler`.

```go
dbHandler := func(db *sql.DB) http.Handler {
    return http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            err := db.Ping()
            // do something with the database here…
            },
        )
        }
http.Handle("/three", dbHandler(db))
```

You create a function that accepts a pointer to a SQL database object and returns a handler, then assign it to a variable named dbHandler. Since `dbHandler` is a function, you can call it with the `db` object to create a handler.

This approach can get a bit cumbersome if you have multiple handlers that require access to the same database object or your design is evolving and you’re likely to require access to additional objects in the future.

A more extensible approach is to use a `struct` whose fields represent objects
and data you want to access in your handler and to define your handlers
as `struct` methods. Injecting dependencies involves adding
struct fields instead of modifying a bunch of closure definitions.

```go

type Handlers struct {
	db  *sql.DB
	log *log.Logger
}

func (h *Handlers) Handler1() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := h.db.Ping()
			if err != nil {
				h.log.Printf("db ping: %v", err)
			}
			// do something with the database here
		},
	)
}
func (h *Handlers) Handler2() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// ...
		},
	)
}

h := &Handlers{
    db: db,
    log: log.New(os.Stderr, "handlers: ", log.Lshortfile),
}
http.Handle("/one", h.Handler1())
http.Handle("/two", h.Handler2())
```

## Middleware

Middleware comprises reusable functions that accept an `http.Handler` and return an `http.Handler`.

You can use middleware to inspect the request and make decisions based on its content before passing it along to the next handler. Or you might use the request content to set headers in the response.

For example, the middleware could respond to the client with an error if the handler
requires authentication and an unauthenticated client sent the request.

Middleware can also collect metrics, log requests, or control access to resources, to name a few uses.

Best of all, you can reuse them on multiple handlers. If you find yourself writing the same handler code over and over, ask yourself if you can put the functionality into middleware and reuse it across your handlers.

check `middleware.go` for more details.

I don’t recommend performing so many tasks in a single middleware function. Instead, it’s best to follow the Unix philosophy and write minimalist middleware, with each function doing one thing very well.

The `net/http` package includes useful middleware to serve static files, redirect requests, and manage request time-outs. Let’s dig into their source code to see how you might use them.


### Timing Out Slow Clients

It’s important not to let clients dictate the duration of a request-response life cycle. Malicious clients could use this leniency to their ends and exhaust your server’s resources, effectively denying service to legit clients. Yet at the same time, setting read and write time-outs server-wide makes it hard for the server to stream data or use different time-out durations for each handler.

Instead, you should manage time-outs in middleware or individual handlers. The `net/http` package includes a middleware function that allows you to control the duration of a request and response on a per-handler basis.

The `http.TimeoutHandler` accepts an `http.Handler`, a duration, and a string to write to the response body. It sets an internal timer to the given duration. If the `http.Handler` does not return before the timer expires, the `http.TimeoutHandler` blocks the `http.Handler` and responds to the client with a `503 Service Unavailable` status.

check `TestTimeoutMiddleware` for more details.

### Protecting Sensitive Files

Middleware can also keep clients from accessing information you’d like to
keep private. For example, the `http.FileServer` function simplifies the process of serving static files to clients, accepting an `http.FileSystem` interface, and returning an `http.Handler`.

The problem is, it won’t protect against serving up potentially sensitive files. Any file in the target directory is fair game.

By convention, many operating systems store configuration files or other sensitive information in files and directories prefixed with a period and hide these dot-prefixed files and directories by default. (This is particularly true in Unix-compatible systems.) But the `http.FileServer` will gladly serve dot-prefixed files or traverse dot-prefixed directories.

Check `RestrictPrefix` in `middleware.go` for more details.

A better approach to restricting access to resources would be to block all resources by default and explicitly allow select resources.

## Multiplexers

A `multiplexer`, like the friendly librarian routing me to the proper bookshelf, is a general handler that routes a request to a specific handler.

The `http.ServeMux` multiplexer is an `http.Handler` that routes an incoming request to the proper handler for the requested resource. By default, `http.ServeMux` responds with a `404 Not Found` status for all incoming requests, but you can use it to register your own patterns and corresponding handlers. It will then compare the request’s URL path with its registered patterns, passing the request and response writer to the handler that corresponds to the longest matching pattern.

check `mux_test.go` for more details.

Go’s multiplexer treats absolute paths as exact matches: either the request’s URL path matches, or it doesn’t. By contrast, it treats subtrees as prefix matches. In other words, the multiplexer will look for the longest registered pattern that comes at the beginning of the request’s URL path. For example, /hello/there/ is a prefix of /hello/there/you but not of /hello/you.

Go’s multiplexer can also redirect a URL path that doesn’t end in a forward slash, such as /hello/there. In those cases, the http.ServeMux first attempts to find a matching absolute path. If that fails, the multiplexer appends a forward slash, making the path /hello/there/, for example, and responds to the client with it. This new path becomes a permanent redirect.

## HTTP/2 Server Pushes

The Go HTTP server can push resources to clients over HTTP/2, a feature that has the potential to improve efficiency.

For example, a client may request the home page from a web server, but the client won’t know it needs the associated style sheet and images to properly render the home page until it receives the HTML in the response.

An HTTP/2 server can proactively send the style sheet and images along with the HTML in the response, saving the client from having to make subsequent calls for those resources.

But server pushes have the potential for abuse.

### Pushing Resources to the Client

Let’s write a command line executable that can push resources to clients.

check `push.go` for more details.

Web browsers cache HTTP/2-pushed resources for the life of the connection and make it available across routes. Therefore, if the index2.html
file served by the /2 route references the same resources pushed by the
default route, and the client first visits the default route, the client’s web
browser may use the pushed resources when rendering the /2 route.

Go doesn’t include the support needed to test the server’s push functionality with code,
but you can interact with this program by using your web browser.


### Don’t Be Too Pushy

Although HTTP/2 server pushing can improve the efficiency of your communications, it can do just the opposite if you aren’t careful. Remember that web browsers store pushed resources in a separate cache for the lifetime of the connection. If you’re serving resources that don’t change often, the web browser will likely already have them in its regular cache, so you shouldn’t push them.

Once it caches them, the browser can use them for future requests spanning many connections. You probably shouldn’t push the resources like (index2.html, style.css, and image.png), for instance, because they’re unlikely to change often.

My advice is to be conservative with server pushes. Use your handlers and rely on metrics to figure out when you should push a resource. If you do push resources, do so before writing the response.
