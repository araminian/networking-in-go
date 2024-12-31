# WRITING HTTP CLIENTS

## Understanding the Basics of HTTP

HTTP is a sessionless client-server protocol in which the client initiates a request to the server and the server responds to the client. HTTP is an application layer protocol and serves as the foundation for communication over the web. It uses TCP as its underlying transport layer protocol.

### Uniform Resource Locators

A URL is an address of sorts used by the client to locate a web server and identify the requested resource.

It’s composed of five parts: a required `scheme` indicating the protocol to use for the connection, an optional `authority` for the resource, the `path` to the resource, an optional `query`, and an optional `fragment`. A colon (:) followed by two forward slashes (//) separates the `scheme` from the `authority`.

The `authority` includes an optional colon-delimited `username` and `password` suffixed with an at symbol (@), a `hostname`, and an optional `port` number preceded by a colon.

The `path` is a series of segments preceded by a forward slash. A question mark (?) indicates the start of the `query`, which is conventionally composed of key-value pairs separated by an `ampersand (&)`.

A hash mark (#) precedes the `fragment`, which is an identifier to a subsection of the resource.

```
scheme://user:password@host:port/path?key1=value1&key2=value2#table_of_contents
```

The typical URL you use over the internet includes a scheme and a hostname at minimum.

```
automobile://the.grocery.store/purchase?butter=irish&eggs=12&coffee=dark_roast
```

This tells me I’m to drive my car to the grocery store and pick up Irish butter, a dozen eggs, and dark roast coffee. It’s important to mention that the scheme is relevant only to the context in which it’s used. My web browser wouldn’t know what to do with the automobile scheme, but for the sake of my marriage, I sure do.

### Client Resource Requests

An HTTP request is a message sent from a client to a web server that asks the server to respond with a specific resource. 

The request consists of a method, a target resource, headers, and a body. 

The `method` tells the server what you want it to do with the target resource. For example, the `GET` method followed by `robots.txt` tells the server you want it to send you the `robots.txt` file, whereas the `DELETE` method indicates to the server that you want it to delete that resource.

Request headers contain metadata about the content of the request you are sending. The `Content-Length` header, for example, specifies the size of the request body in bytes. 

The request body is the payload of the request. If you upload a new profile picture to a web server, the request body will contain the image encoded in a format suitable for transport over the network, and the `Content-Length` header’s value will be set to the size in bytes of the image in the request body. Not all request methods require a request body.

```bash
$ nc www.google.com 80
GET /robots.txt HTTP/1.1
```


`GET` As in the earlier example, the `GET` method instructs the server to send you the target resource. The server will deliver the target resource in the response’s body. It’s important to note that the target resource does not need to be a file; the response could deliver you dynamically generated content, like the gophers image search result discussed earlier. The server should never change or remove the resource as the result of a GET request.

`HEAD` The `HEAD` method is like `GET` except it tells the server to exclude the target resource in its response. The server will send only the response code and other various bits of metadata stored in the response headers. You can use this method to retrieve meaningful details about a resource, such as its size, to determine whether you want to retrieve the resource in the first place. (The resource may be larger than you expect.)

`POST` A `POST` request is a way for you to upload data included in the request body to a web server. The `POST` method tells the server that you are sending data to associate with the target resource. For example, you may post a new comment to a news story, in which case the news story would be the target resource. In simple terms, think of the `POST` method as the method for creating new resources on the server.

`PUT` Like POST, you can use a PUT request to upload data to a web server. In practice, the PUT method usually updates or completely replaces an existing resource. You could use PUT to edit the comment you POSTed to the news story.

`PATCH` The `PATCH` method specifies partial changes to an existing resource, leaving the rest of the resource unmodified. In this way, it’s like a diff. Let’s assume you are buying a Gopher Plush for that someone special in your life, and you’ve proceeded past the shipping address step of the checkout process when you realize you made a typo in your street address. You jump back to the shipping address form and correct the typo. Now, you could POST the form again and send all its contents to the server. But a PATCH request would be more efficient since you made only a single correction. You’ll likely encounter the PATCH method in APIs, rather than HTML forms.

`DELETE` The `DELETE` method instructs the server to remove the target resource. Let’s say your comment on the news story was too controversial, and now your neighbors avoid making eye contact with you. You can make a DELETE request to the server to remove your comment and restore your social status.

`OPTIONS` You can ask the server what methods a target resource supports by using the OPTIONS method. For example, you could send an OPTIONS request with your news story comment as the target resource and learn that the DELETE method is not one of the methods the server will support for your comment, meaning your best option now is to find another place to live and meet new neighbors.

`CONNECT` The client uses CONNECT to request that the web server perform HTTP tunneling, or establish a TCP session with a target destination and proxy data between the client and the destination.

`TRACE` The `TRACE` method instructs the web server to echo the request back to you instead of processing it. This method allows you to see whether any intermediate nodes modify your request before it reaches the web server.

Before adding server-side support for the `TRACE` method, I strongly recommend you read up on its role in cross-site tracing (XST) attacks, whereby an attacker uses a cross-site scripting (XSS) attack to steal authenticated user credentials. The risk of adding a potential attack vector to your web server likely does not outweigh the diagnostic benefits of `TRACE` support.

`TRACE` using nc:

```bash
$ nc www.google.com 80
TRACE / HTTP/1.1
```
using `curl`:
```bash
$ curl -X TRACE www.google.com
```

### Server Responses

Whereas the client request always specifies a method and a target resource, the web server’s response always includes a status code to inform the client of the status of its request. 

The status code is a three-digit number that begins with a 1, 2, 3, 4, or 5. The first digit indicates the class of the status code, and the last two digits indicate the specific status code.

A successful request results in a response containing a `200`-class status code.

If the client makes a request that requires further action on the client’s part, the server will return a `300`-class status code. For example, if the client requests a resource that has not changed since the client’s last request for the resource, the server may return a `304` status code to inform the client that it should instead render the resource from its cache.

If an error occurs because of the client’s request, the server will return a `400`-class status code in its response. The most common example of this scenario occurs when a client requests a nonexistent target resource, in which case the server responds with a `404` status code to inform the client that it could not find the resource.

The `500`-class status codes inform the client that an error has occurred on the server side that prevents the server from fulfilling the request. Let’s assume that your request requires the web server to retrieve assets from an upstream server to satisfy your request, but that the upstream server fails to respond. The web server will respond to you with a `504` status code indicating that a time-out occurred during the communication with its upstream server.

A handful of `100`-class status codes exist in HTTP/1.1 to give direction to the client. For example, the client can ask for guidance from the server while sending a `POST` request. To do so, the client would send the `POST` method, target resource, and request headers to the server, one of which tells the server that the client wants permission to proceed sending the request body. The server can respond with a `100` status code indicating that the client can continue the request and send the body.

Go defines many of these status codes as constants in its `net/http` package, and I suggest you use the constants in your code. 

`200 OK` Indicates a successful request. If the request method was `GET`, the response body contains the target resource.

`201 Created` Returned when the server has successfully processed a request and added a new resource, as may be the case with a `POST` request.

`202 Accepted` Often returned if the request was successful but the server hasn’t yet created a new resource. The creation of the resource may still fail despite the successful request.

`204 No Content` Often returned if the request was successful but the response body is empty.

`304 Not Modified` Returned when a client requests an unchanged resource. The client should instead use its cached copy of the resource. One method of caching is using an `entity tag (ETag) header`. When a client requests a resource from the server, the response may include an optional server-derived `ETag header`, which has meaning to the server. If the client requests the same resource in the future, the client can pass along the cached `ETag header` and value in its request. The server will check the `ETag` value in the client’s request to determine whether the requested resource has changed. If it is unchanged, the server will likely respond with a `304` status code and an empty response body.

`400 Bad Request` Returned if the server outright rejects the client’s request for some reason. This may be due to a malformed request, like one that includes a request method but no target resource.

`403 Forbidden` Often returned if the server accepts your request but determines you do not have permission to access the resource, or if the server itself does not have permission to access the requested resource.

`404 Not Found` Returned if you request a nonexistent resource. You may also find this status code used as a Glomar response when a server does not want to confirm or deny your permission to access a resource. In other words, a web server may respond with a 404 status code for a resource you do not have permission to access instead of explicitly responding with a 403 status code confirming your lack of permissions to the resource. Attackers attempting to access sensitive resources on your web server would want to focus their efforts only on existing resources, even if they currently lack permissions to access those resources. Returning a 404 status code for both nonexistent and forbidden resources prevents attackers from differentiating between the two, providing a measure of security. The downside to this approach is you’ll have a harder time debugging your permissions on your server, because you won’t know whether the resource you’re requesting exists or you simply lack permissions. I suggest you articulate the difference in your server logs.

`405 Method Not Allowed` Returned if you specify a request method for a target resource that the server does not support. Remember the controversial comment you attempted to delete in our discussion of the `OPTIONS` request method? You would receive a `405` status code in response to that `DELETE` request.

`426 Upgrade Required` Returned to instruct the client to first upgrade to TLS before requesting the target resource.

`500 Internal Server Error` A catchall code of sorts returned when an error occurs on the server that prevents it from satisfying the client’s request but that doesn’t match the criteria of any other status code. Servers have returned many a `500` error because of some sort of configuration error or syntax error in server-side code. If your server returns this code, check your logs.

`502 Bad Gateway` Returned when the server proxies data between the client and an upstream service, but the upstream service is unavailable and not accepting requests.

`503 Service Unavailable` Returned when the server is temporarily unable to process the request due to a temporary overloading or maintenance of the server. The server may be down for maintenance or experiencing high traffic.

`504 Gateway Timeout` Returned by a proxy web server to indicate that the upstream service accepted the request but did not provide a timely reply.

### From Request to Rendered Page

In `HTTP version 1.0 (HTTP/1.0)`, clients must initiate a separate TCP connection for each request. `HTTP/1.1` eliminates this requirement, reducing the latency and request-connection overhead associated with multiple HTTP requests to the same web server. Instead, it allows multiple requests and responses over the same TCP connection. 

The latest version of HTTP, `HTTP/2`, aims to further reduce latency. In addition to reusing the same TCP connection for subsequent requests, the `HTTP/2` server can proactively push resources to the client. The client requested the default resource. The server responded with the default resource. But since the server knew that the default resource had dependent resources, it pushed those resources to the client without the client’s needing to make separate requests for each resource. 

## Retrieving Web Resources in Go

Go won’t directly render an HTML page to your screen. Instead, you could use Go to scrape data from websites (such as financial stock details), submit form data, or interact with APIs that use HTTP for their application protocol, to name a few examples.

### Using Go’s Default HTTP Client

The net/http package includes a default client that allows you to make one off HTTP requests.

```go
resp, err := http.Get("https://www.google.com")
```

`TestHeadTime` demonstrates one way you can retrieve the current time from a trusted authority—time.gov’s web server—and compare it with the local time on your computer.

The `net/http` package includes a few helper functions to make `GET, HEAD, or POST` requests. 


### Closing the Response Body

As mentioned earlier, `HTTP/1.1` allows the client to maintain a TCP connection with a server for multiple HTTP requests (we call this keepalive support). 

Even so, the client **cannot reuse a TCP session when unread bytes from the previous response remain on the wire**. 

Go’s HTTP client **automatically drains the response body when you close it**. This allows your code to reuse the underlying TCP session if you are diligent about closing every response body.

The client doesn’t automatically read the response body, however. The body remains unconsumed until your code explicitly reads it or until you close it and Go implicitly drains any unread bytes.

**NOTE**:
The Go HTTP client’s implicit draining of the response body on closing could potentially bite you. For example, let’s assume you send a GET request for a file and receive a response from the server. You read the response’s `Content-Length` header and realize the file is much larger than you anticipated. If you close the response body without reading any of its bytes, Go will download the entire file from the server as it drains the body regardless.

A better alternative would be to send a `HEAD` request to retrieve the `ContentLength` header. This way, no unread bytes exist in the response body, so closing the response body will not incur any additional overhead while draining it.

On the rare occasion that you make an HTTP request and want to **explicitly drain the response body**, the most efficient way is to use the `io.Copy` function:

```go
io.Copy(io.Discard, resp.Body)
resp.Close()
```

The `io.Discard` is a `io.Writer` that discards all data written to it.

### Implementing Time-outs and Cancellations

Go’s default HTTP client and the requests created with the `http.Get`, `http.Head`, and `http.Post` helper functions do not time out.

The consequences of this may not be obvious until you get bit by the following fact (after which you’ll never forget it): the lack of a time-out or deadline means that a **misbehaving or malicious service could cause your code to block indefinitely without producing an error to indicate that anything’s wrong**. You might not find out that your service is malfunctioning until users start calling you to complain.

`TestBlockIndefinitely`: demonstrates a simple test that causes the HTTP client to block indefinitely.

To solve this issue, production code should use the technique you learned for timing out network sockets in “Using a Context with a Deadline to Time Out a Connection”.

Create a `context` and use it to initialize a new request. You can then manually cancel the request by either using the context’s cancel function or creating a context with a deadline or time-out.

Check out `TestBlockIndefinitelyWithTimeout` : time out the request after five seconds without an answer from the server.

**Keep in mind that the context’s timer starts running as soon as you initialize the context. The context controls the entire life cycle of the request. In other words, the client has five seconds to connect to the web server, send the request, read the response headers, and pass the response to your code.**

If you are in the middle of reading the response body when the context times out, your next read will immediately return an error. So, use generous time-out values for your specific application.

Alternatively, create a context without a time-out or deadline and control the cancellation of the context exclusively by using a timer and the context’s cancel function, like this:

```go
ctx, cancel := context.WithCancel(context.Background())
timer := time.AfterFunc(5*time.Second, cancel)
// Make the HTTP request, read the response headers, etc.
// ...
// Add 5 more seconds before reading the response body.
timer.Reset(5*time.Second)

```

This snippet demonstrates how to use a timer that will call the context’s `cancel` function after it expires. You can reset the timer as needed to push the call to cancel further into the future.

### Disabling Persistent TCP Connections

By default, Go’s HTTP client maintains the underlying TCP connection to a web server after reading its response unless explicitly told to disconnect by the server.

Although this is desirable behavior for most use cases because it allows you to use the same TCP connection for multiple requests, you may inadvertently deny your computer the ability to open new TCP connections with other web servers.

This is because the number of active TCP connections a computer can maintain is finite. If you write a program that makes one-off requests to numerous web servers, you could find that your program stops working after exhausting all your computer’s available TCP connections, leaving it unable to open new ones.

In this scenario, TCP session reuse can work against you. Instead of disabling TCP session reuse in the client, a more flexible option is to inform the client what to do with the TCP socket on a per-request basis.

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
if err != nil {
t.Fatal(err)
}
req.Close = true
```

Setting the request’s `Close` field to `true` tells Go’s HTTP client that it should close the underlying TCP connection after reading the web server’s response. If you know you’re going to send four requests to a web server and no more, you could set the `Close` field to `true` on the fourth request. All four requests will use the same TCP session, and the client will terminate the TCP connection after receiving the fourth response.

## Posting Data over HTTP

This payload can be any object that implements the `io.Reader` interface, including a file handle, standard input, an HTTP response body, or a Unix domain socket, to name a few. But as you’ll see, sending data to the web server involves a little more code than a GET request because you must prepare that request body.


### Posting JSON to a Web Server

Before you can send data to a test server, you need to create a handler that can accept it.
`server_test.go` : creates a new type named `User` that you will encode to JavaScript Object Notation (JSON) and post to the handler.


