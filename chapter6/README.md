# ENSURING UDP RELIABILITY

## Reliable File Transfers Using TFTP

It’s your application’s job to make the UDP connection reliable.

The `Trivial File Transfer Protocol (TFTP)` is an example of an application protocol that ensures reliable data transfers over UDP. It allows two nodes to transfer files over UDP by implementing a subset of the features that make TCP reliable.

A `TFTP server` implements ordered packet delivery, acknowledgments, and retransmissions. 

Please keep in mind that TFTP is not appropriate for secure file transmission. Though it adds reliability to UDP connections, it does not support encryption or authentication. 

## TFTP Types

Your TFTP server will accept read requests from the client, send data packets, transmit error packets, and accept acknowledgments from the client.

check `tftp.go` for the full implementation.

The first `2 bytes` of a TFTP packet’s header is an operation code.
As with the operation codes, you define a series of unsigned 16-bit integer error codes.

### READ REQUEST

The server receives a `read request` packet when the client wants to download a file. 

The server must then respond with either a `data packet` or an `error packet`.

Either packet serves as an acknowledgment to the client that the server received the read request.
If the client does not receive a data or error packet, it may retransmit the read request until the server responds or the client gives up.

Read request packet structure:
- 2 bytes: OpRRQ
- n bytes: filename
- 1 byte: 0 (null terminator)
- n bytes: mode
- 1 byte: 0 (null terminator)

The mode indicates to the server how it should send the file: `netascii` or `octet`. If a client requests a file using the `netascii` mode, the client must convert the file to match its own line-ending format. For our purposes, you will accept only the `octet` mode, which tells the server to send the file in a binary format, or as is.

Your TFTP server’s read request, data, acknowledgment, and error packets all implement the `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces. 

### DATA PACKET

Clients receive data packets in response to their read requests, provided the server was able to retrieve the requested file. The server sends the file in a series of data packets, each of which has an assigned block number, starting at 1 and incrementing with every subsequent data packet. 

The block number allows the client to properly order the received data and account for duplicates.

All data packets have a payload of 512 bytes except for the last packet. The client continues to read data packets until it receives a data packet whose payload is less than 512 bytes, indicating the end of the transmission. 

At any point, the client can return an error packet in place of an acknowledgment, and the server can return an error packet instead of a data packet. An error packet immediately terminates the transfer.

Data packet structure:
- 2 bytes: OpDATA
- 2 bytes: block number
- n bytes: data (up to 512 bytes)

The server requires an acknowledgment from the client after each data packet. If the server does not receive a timely acknowledgment or an error from the client, the server will retry the transmission until it receives a reply or exhausts its number of retries.

Once the client has sent the initial read request packet, the server responds with the first block of data. Next, the client acknowledges receipt of block 1. The server receives the acknowledgment and replies with the second block of data.

You may have recognized the potential for an integer overflow of the 16-bit, unsigned block number. If you send a payload larger than about 33.5MB (65,535 × 512 bytes), the block number will overflow back to 0.

Your server will happily continue sending data packets, but the client may not be as graceful handling the overflow. You should consider mitigating overflow risks by limiting the file size the TFTP server will support so as not to trigger the overflow, recognizing that an overflow can occur and determining whether it is acceptable to the client, or using a different protocol altogether.


### ACKNOWLEDGMENT PACKET

The client uses the block number to send a corresponding acknowledgment to the server and to properly order this block of data among the other received blocks of data.

Acknowledgment packet structure:
- 2 bytes: OpACK
- 2 bytes: block number

### Handling Errors

Clients and servers convey errors by using an `error packet`.

Error packet structure:
- 2 bytes: OpERROR
- 2 bytes: error code
- n bytes: error message
- 1 byte: 0 (null terminator)

## TFTP Server

check `server.go`.

The fact that your packet types implement the `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces means that your server code can act as a conduit between the network interface and these types, leading to simpler code. All your server must concern itself with is transferring byte slices between your types and the network connection.


### Handle Read Request

The handler accepts read requests from the client and replies with the `server’s payload`.

The handler sends one data packet and waits for an acknowledgment from the client before sending another data packet. It also attempts to retransmit the current data packet when it fails to receive a timely reply from the client.



