## UDP

UDP is unreliable because it does not include many of the mechanisms that make TCP so trustworthy. It provides little more than a socket address (an IP address and port).

Unlike TCP, UDP does not provide session support or even confirm that the destination is accessible; it simply makes a best-effort attempt to deliver the packet.

UDP does not manage congestion, control data flow, or retransmit packets. Lastly, UDP does not guarantee that the destination receives packets in the order they originate. UDP is simply a conduit between applications and the IP layer.

UDP has a few strengths over TCP. Whereas TCP must establish a session with each individual node in a group before it can transmit data, UDP can send a single packet to a group of nodes without duplicating the packet, a process known as multicasting.

UDP is also capable of broadcasting packets to all members of a subnet since it doesn’t need to establish a session between each node.

UDP is ideal when missing packets aren’t detrimental to overall communication because the most recently received packets can take the place of earlier, dropped packets.

You should consider using UDP in your application if it doesn’t require all the features TCP provides. For most network applications, TCP is the right protocol choice. But UDP is an option if its speed and simplicity better fit your use case and the reliability trade-offs aren’t detrimental.

UDP’s packet structure consists of an `8-byte header` and a `payload`. The header contains `2 bytes for the source port`, `2 bytes for the destination port`, `2 bytes for the packet length in bytes`, and a `2-byte checksum`. The minimum packet length is `8 bytes` to account for the header and an empty payload.

Although the maximum packet length is 65,535 bytes, application layer protocols often limit the packet length to avoid fragmentation.

UDP is a connectionless protocol, which means that it does not establish a connection before sending packets. Instead, it sends packets directly to the destination address.

UDP is a stateless protocol, which means that it does not maintain any state information about the connection.

When it comes to sending and receiving data, UDP is uncivilized compared to TCP.

the `net.Conn` interface for handling stream-oriented connections, such as TCP, between a client and a server. But this interface isn’t ideal for UDP connections because UDP is not a stream-oriented protocol. UDP does not maintain a session or involve a handshake process like TCP. UDP does not have the concept of acknowledgments, retransmissions, or flow control.

Instead, UDP primarily relies on the packet-oriented `net.PacketConn` interface.

## UDP Echo Server

Sending and receiving UDP packets is a nearly identical process to sending and receiving TCP packets. But since UDP doesn’t have session support, you must be able to handle an additional return value, the sender’s address, when reading data from the connection object.

check out the `echo.go` file for the full implementation.

Notice there is no Accept method on your UDP connection as there is with the TCP-based listeners in the previous chapters. This is because UDP doesn’t use a handshake process.

To write a UDP packet, you pass a byte slice and a destination address to the connection’s `WriteTo` method.

To read a UDP packet, you pass a byte slice to the connection’s `ReadFrom` method. This method returns the number of bytes read and the sender’s address.

## Every UDP Connection Is a Listener

The `net.PacketConn` interface is a listener for UDP packets.

The `net.ListenPacket` function creates a `net.PacketConn` object.

The `net.PacketConn` interface has a `ReadFrom` method that reads a UDP packet from the connection.

The `net.PacketConn` interface has a `WriteTo` method that writes a UDP packet to the connection.

`net` package distinguishes between a TCP connection object (`TCPConn`) and a TCP listener (`TCPListener`). The TCP listener is what accepts the connection and returns an object that represents
the listener’s side of the connection so that the listener can then send a message to the client.

There is no UDP equivalent of the `TCPListener` because UDP lacks sessions. You need to verify the sender’s address, because you can no longer trust that all incoming packets to a connection object are from the same sender.

## Using net.Conn in UDP

You can establish a UDP connection that implements the `net.Conn` interface so that your code behaves indistinguishably from a TCP `net.Conn`.

You do so by passing `udp` as the first argument to the `net.Dial` function.

Using `net.Conn` with your UDP-based connections can **prevent interlopers from sending you messages** and **eliminate the need to check the sender’s address on every reply you receive**.

check out the `echo_test.go` file for the full implementation.

The client side of a connection can leverage the stream-oriented functionality of `net.Conn` over UDP, but the UDP listener must still use `net.PacketConn`.

Unlike TCP, the echo server receives no traffic upon calling `net.Dial` because no handshake is necessary.

For your purposes, using net.Conn over net.PacketConn may make your UDP connection code cleaner. Just be aware of the trade-offs. Using net.Conn with UDP does not offer the same functionality as you would expect when using net.Conn with TCP. For example, a UDP-based net.Conn’s Write method will not return an error if the destination failed to receive the packet. The onus is still on your application code to confirm delivery when using UDP.
