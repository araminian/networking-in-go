# SECURING COMMUNICATION WITH TLS

## A Closer Look at Transport Layer Security

The TLS protocol supplies secure communication between a client and a server. It allows the client to authenticate the server and optionally permits the server to authenticate clients.

The client uses TLS to encrypt its communication with the server, preventing third-party interception and manipulation.

TLS uses a handshake process to establish certain criteria for the stateful TLS session. If the client initiated a TLS 1.3 handshake with the server, it would go something like this:

Client: Hello google.com. I’d like to communicate with you using TLS version 1.3. Here is a list of ciphers I’d like to use to encrypt our messages, in order of my preference. I generated a public- and private-key pair specifically for this conversation. Here’s my public key.    

Server: Greetings, client. TLS version 1.3 suits me fine. Based on your cipher list, I’ve decided we’ll use the `Advanced Encryption Standard with Galois/Counter Mode (AES-GCM) cipher`. I, too, created a new key pair for `this conversation`. Here is `my public key` and `my certificate`q so you can prove that I truly am google.com. I’m also sending along a 32-byte value that corresponds to the TLS version you asked me to use. Finally, I’m including both a `signature and a message authentication code (MAC)` derived using your public key of everything we’ve discussed so far so you can verify the integrity of my reply when you receive it.

Client (to self): An authority I trust signed the server’s certificate, so
I’m confident I’m speaking to google.com. I’ve derived this conversation’s `symmetric key from the server’s signature by using my private key`. Using this symmetric key, I’ve verified the MAC and made sure no one has tampered with the server’s reply. The 32 bytes in the reply corresponds to TLS version 1.3, so no one is attempting to trick the server into using an older, weaker version of TLS. I now have everything I need to securely communicate with the server.

Client: (to server) Here is some encrypted data.

The 32-byte value in the server’s hello message prevents downgrade attacks, in which an attacker intercepts the client’s hello message and modifies it to request an older, weaker version of TLS.

From this point forward, the client and server communicate using `AES-GCM symmetric-key cryptography` (in this hypothetical example).
Both the client and the server encapsulate application layer payloads in TLS records before passing the payloads onto the transport layer.

Once the payload reaches its destination, TLS receives the payload from the transport layer, decrypts it, and passes the payload along to the application layer protocol.

## Forward Secrecy

The handshake method in our hypothetical conversation is an example of the Diffie-Hellman (DH) key exchange used in TLS v1.3. 

The DH key exchange calls for the creation of new client and server key pairs, and a new symmetric key, all of which should exist for only the duration of the session.
Once a session ends, the client and server shall discard the session keys.

The use of per-session keys means that TLS v1.3 gives you forward secrecy;
an attacker who compromises your session keys can compromise only the data exchanged during that session. An attacker cannot use those keys to decrypt data exchanged during any other session.

## In Certificate Authorities We Trust

LS’s certificates work in much the same way as my passport. If I wanted a new TLS certificate for woodbeck.net, I would send a request to a certificate authority, such as Let’s Encrypt. The certificate authority would then verify I am the proper owner of woodbeck.net. Once satisfied, the certificate authority would issue a new certificate for woodbeck.net and cryptographically sign it with its certificate. My server can present this certificate to clients so they can authenticate my server by confirming the certificate authority’s signature, giving them the confidence that they’re communicating with the real woodbeck.net, not an impostor.

A certificate authority issuing a signed certificate for woodbeck.net is analogous to the US government issuing my passport. They are both issued by trusted institutions that attest to their subject’s authenticity. Like Ireland’s trust of the United States, clients are inclined to trust the woodbeck.net certificate only if they trust the certificate authority that signed it. I could create my own certificate authority and self-sign certificates as easy as I could create a document claiming to be my passport. But Ireland would sooner admit that Jack Daniel’s Tennessee Whiskey is superior to Jameson Irish Whiskey than trust my self-issued passport, and no operating system or web browser in the world would trust my self-signed certificate.

## How to Compromise TLS

If an attacker were able to install his own CA certificate in your operating system’s trusted certificate storage, your computer would trust any certificate he signs. This means an attacker could compromise all your TLS traffic.

Most people don’t get this kind of special attention. Instead, an attacker is more likely to compromise a server. Once compromised, the attacker could capture all TLS traffic and the corresponding session keys from memory.

## Protecting Data in Transit

Ensuring the integrity of the data you transmit over a network should be your primary focus, no matter whether it’s your own data or the data of others. Go makes using TLS so easy that you would have a tough time justifying not using it. 

### Client-side TLS

The client’s primary concern during the handshake `process is to authenticate the server by using its certificate`. If the client cannot trust the server, it cannot consider its communication with the server secure.

check out `TestClientTLS` in `client_test.go`.

### TLS over TCP

TLS is stateful; a client and a server negotiate session parameters during the initial handshake only, and once they’ve agreed, they exchange encrypted TLS records for the duration of the session. Since TCP is also stateful, it’s the ideal transport layer protocol with which to implement TLS, because you can leverage TCP’s reliability guarantees to maintain your TLS sessions.

check out `TestClientTLSGoogle` in `client_test.go`.

It demonstrates how to use the `crypto/tls` package to initiate a TLS connection with a few lines of code.

### Server-side TLS

The server-side code isn’t much different from what you’ve learned thus far. The main difference is that the server needs to present a certificate to the client as part of the handshake process. 

```bash
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -host localhost -ecdsa-curve P256
```

This command creates a certificate named cert.pem with the hostname localhost and a private key named key.pem.

check out `server.go` which includes the first bit of code for a TLS-only echo server.


