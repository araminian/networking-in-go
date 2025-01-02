# DATA SERIALIZATION

These services must communicate by exchanging bytes of data in a way that is meaningful to both the sender and receiver, despite the different programming languages they’re using.

To do this, the sender converts data into bytes using a standard format and transfers the bytes over the network to the receiver. If the receiver understands the format used by the sender, it can convert the bytes back into structured data.

The process of converting data into bytes is called serialization. The process of converting bytes back into data is called deserialization.

Services can use data serialization to convert structured data to a series of bytes suitable for transport over a network or for persistence to storage.

## Serializing Objects

Objects or structured data cannot traverse a network connection as is. In other words, you cannot pass in an object to net.Conn’s Write method, since it accepts only a byte slice. Therefore, you need to serialize the object to a byte slice, which you can then pass to the Write method.

Go’s standard library includes excellent support for popular data serialization formats in its encoding package.

You’ve already used `encoding/binary` to serialize numbers into byte sequences, `encoding/json` to serialize objects into JSON for submission over HTTP, and `encoding/pem` to serialize TLS certificates and private keys to files.

This section will build an application that serializes data into three binary encoding formats: JSON, protocol buffers, and Gob.

check out the `homework.go` file for the code.


Go’s JSON and Gob encoding packages can serialize `exported struct fields only`, so you define Chore as a struct, making sure to export its fields.

You could use struct tags to instruct the encoders on how to treat each field, if necessary. For example, you could place the struct tag `json:"-"` on the Complete field to tell Go’s JSON encoder to ignore this field instead of encoding it.

### JSON

JSON is a common, human-readable, text-based data serialization format that uses key-value pairs and arrays to represent complex data structures.

JSON’s types include strings, Booleans, numbers, arrays, key-value objects, and nil values specified by the keyword null. JSON numbers do not differentiate between floating-points and integers.

Check out the `json/homework.go` file for JSON storage implementation using Go’s
encoding/json package.

```bash
go run . add Mop floors, Clean dishes, Mow the lawn
go run . complete 2
go run . list
```

### Gob

Gob, as in “gobs of binary data,” is a `binary serialization` format native to Go. Engineers on the Go team developed Gob to combine the efficiency of protocol buffers, arguably the most popular binary serialization format, with JSON’s ease of use.

For example, protocol buffers don’t let us simply instantiate a new encoder and throw data structures at it, as you did in the JSON.

On the other hand, Gob functions much the same way as the JSON encoder, in that Gob can intelligently infer an object’s structure and serialize it.

If you are communicating with other Go services that support Gob, I recommend you use Gob over JSON. Go’s encoding/gob is more performant than encoding/json. Gob encoding is also more succinct, in that Gob uses less data to represent objects than JSON does. This can make a difference when storing or transmitting serialized data.

### Protocol Buffers

Like Gob, protocol buffers use binary encoding to store or exchange information across various platforms. It’s faster and more succinct than Go’s JSON encoding. Unlike Gob and like JSON, protocol buffers are language neutral and enjoy wide support among popular programming languages.

This makes them ideally suited for using in Go-based services that you hope to integrate with services written in other programming languages.

Protocol buffers use a definition file, conventionally suffixed with .proto, to define messages. Messages describe the structured data you want to serialize for storage or transmission.

```
message Chore {
bool complete = 1;
string description = 2;
}
```

The field’s type and number identify the field in the binary payload, so these must not change once used or you’ll break backward compatibility. However, it’s fine to add new messages and message fields to an existing `.proto` file.

It’s a good practice to treat your protocol buffer definitions as you would an API. If third parties use your protocol buffer definition to communicate with your service, consider versioning your definition; this allows you to create new versions anytime you need to break backward compatibility.

You’ll have to compile the .proto file to generate Go code from it. This code allows you to serialize and deserialize the messages defined in the .proto file. Third parties that want to exchange messages with you can use the same .proto file to generate code for their target programming language.

## Transmitting Serialized Objects

Large technology companies facilitate this with remote procedure calls (RPCs), a technique by which a client can transparently call a subroutine on a server as if the call were local.

From your application’s perspective, RPC services take code that appears to run locally and distribute it over a network. Your code may call a function that transparently relays a message to a server. The server would locally execute the function, then respond with the results, which your code receives as the function’s return value.

As far as your code is concerned, the function call is local, despite RPC’s transparent interaction with the server. This approach allows you to scale services across servers while abstracting the details away from your code. In other words, your code functions the same no matter whether the function call runs on the same computer or on one over the network.

Most companies now implement RPC with gRPC, a cross-platform framework that leverages HTTP/2 and protocol buffers. Let’s use it here to build something more sophisticated than an app to keep track of the housework you have yet to do. You’ll write a service that can send tasks to Rosie, the robotic maid from the classic animated series The Jetsons, who can take over your domestic responsibilities. Granted, she won’t be available until the year 2062, but you can get a head start on the code.

### Connecting Services with gRPC

The gRPC framework is a collection of libraries that abstracts many of the RPC implementation details. It is platform neutral and programminglanguage agnostic; you can use it to integrate a Go service running on Windows with a Rust service running on Linux, for example.

### Creating a TLS-Enabled gRPC Server

Now let’s implement a gRPC client and server. By default, gRPC requires a secure connection, so you’ll add TLS support to your server.

We need to implement the `RobotMaid` service in the `homework/v1/homework.proto` file.

```go
type RobotMaidServer interface {
Add(context.Context, *Chores) (*Response, error)
Complete(context.Context, *CompleteRequest) (*Response, error)
List(context.Context, *empty.Empty) (*Chores, error)
mustEmbedUnimplementedRobotMaidServer()
}
```

We need to implement the `RobotMaidServer` interface in the `server/robot.go` file.


### Creating a gRPC Client to Test the Server

The client-side code isn’t much different from what you wrote in the “Serializing Objects” section on page 270. The main difference is that you need to instantiate a new gRPC client and modify the add, complete, and list functions to use it.

check out the `client/client.go` file for the client-side code.

