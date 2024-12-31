# UNIX DOMAIN SOCKETS

Your applications may sometimes need to communicate with services, such as a database, hosted on the same node.

One way to connect your application to a database running on the same system would be to send data to the node’s IP address or localhost address—commonly 127.0.0.1—and the database’s port number. However, there’s another way: using Unix domain sockets. The Unix domain socket is a communication method that uses the filesystem to determine a packet’s destination address, allowing services running on the same node to exchange data with one another, a process known as `inter-process communication (IPC)`.

## What Are Unix Domain Sockets?

Socket addressing allows individual services on the same node, at the same IP address, to listen for incoming traffic. the IP address and port number of a socket address allow you to communicate with every single service listening to each socket address on a node.

Unix domain sockets apply the socket-addressing principle to the filesystem: each Unix domain socket has an associated file on the filesystem, which corresponds to a network socket’s IP address and port number. You can communicate with a service listening to the socket by reading from and writing to this file.

Likewise, you can leverage the filesystem’s ownership and permissions to control read and write access to the socket. 

Unix domain sockets increase efficiency by bypassing the operating system’s network stack, eliminating the overhead of traffic routing. For the same reasons, you won’t need to worry about fragmentation or packet ordering when using Unix domain sockets. If you choose to forgo Unix domain sockets and exclusively use network sockets when communicating with local services (for example, to connect your application to a local database, a memory cache, and so on), you ignore significant security advantages and performance gains.

Though this system brings distinct advantages, it comes with a caveat: Unix domain sockets are local to the node using them, so you cannot use them to communicate with other nodes, as you can with network sockets. Therefore, Unix domain sockets may not be a good fit if you anticipate moving a service to another node or require maximum portability for your application. To maintain communication, you’d have to first migrate to a network socket.

## Binding to Unix Domain Socket Files

A Unix domain socket file is created when your code attempts to bind to an unused Unix domain socket address by using the `net.Listen`, `net.ListenUnix`, or `net.ListenPacket` functions.

If the socket file for that address already exists, the operating system will return an error indicating that the address is in use.

In most cases, simply removing the existing Unix domain socket file is enough to clear up the error. However, you should first make sure that the socket file exists not because a process is currently using that address but because you didn’t properly clean up the file from a defunct process.

If you wish to reuse a socket file, use the `net` package’s `FileListener` function to bind to an existing socket file.

Once a service binds to the socket file, you can use Go’s `os` package to modify the file’s ownership and read/write permissions. Specifically, the `os.Chown` function allows you to modify the user and group that owns the file.

```go
err := os.Chown("/path/to/socket/file", -1, 100) // owner ID , group ID
```

A user or group ID of -1 tells Go you want to maintain the current user or group ID.

```go
// lookup group
grp, err := user.LookupGroup("users")
```

```go
// set permissions 
// os.ModeSocket is a bitmask that indicates the file is a socket
err := os.Chmod("/path/to/socket/file", os.ModeSocket|0660)
```

## Understanding Unix Domain Socket Types
There are three types of Unix domain sockets: `streaming sockets`, which operate like TCP; `datagram sockets`, which operate like UDP; and `sequence packet sockets`, which combine elements of both.

Go designates these types as `unix`, `unixgram`, and `unixpacket`, respectively.

The `net.Conn` interface allows you to write code once and use it across multiple network types. It abstracts many of the differences between the network sockets used by TCP and UDP and Unix domain sockets, which means that you can take code written for communication over TCP, for example, and use it over a Unix domain socket by simply changing the address and network type.

### The unix Streaming Socket

The streaming Unix domain socket works like TCP without the overhead associated with TCP’s acknowledgments, checksums, flow control, and so on. The operating system is responsible for implementing the streaming inter-process communication over Unix domain sockets in lieu of TCP.

A listener created with either `net.Listen` or `net.ListenUnix` will automatically remove the socket file when the listener exits.

Unix domain socket files created with `net.ListenPacket` won’t be automatically removed when the listener exits.

check `TestEchoServerUnix`.

### The unixgram Datagram Socket

Next let’s create an echo server that will communicate using datagrambased network types, such as udp and unixgram. 

Whether you’re communicating over UDP or a unixgram socket, the server you’ll write looks essentially the same. The difference is, you will need to clean up the socket file with a unixgram listener.


check `TestEchoServerUnixDatagram`.