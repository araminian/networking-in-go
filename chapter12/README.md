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