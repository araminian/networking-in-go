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

### Certificate Pinning

Earlier in the chapter, we discussed ways to compromise the trust that TLS relies on, whether by a certificate authority issuing fraudulent certificates or an attacker injecting a malicious certificate into your computer’s trusted certificate storage. You can mitigate both attacks by using certificate pinning.

Certificate pinning is the process of scrapping the use of the operating system’s trusted certificate storage and explicitly defining one or more trusted certificates in your application. Your application will trust connections only from hosts presenting a pinned certificate or a certificate signed by a pinned certificate.

If you plan on deploying clients in zero-trust environments that must securely communicate with your server, consider pinning your server’s certificate to each client.

Assuming the server introduced in the preceding section uses the cert.pem and the key.pem you generated for the hostname localhost, all clients will abort the TLS connection as soon as the server presents its certificate. Clients won’t trust the server’s certificate because no trusted certificate authority signed it. You could set the tls.Config’s InsecureSkipVerify field to true, but as this method is insecure, I don’t recommend you consider it a practical choice.

Instead, let’s explicitly tell our client it can trust the server’s certificate by pinning the server’s certificate to the client.

check `TestEchoServerTLS` in `server_test.go`.

## Mutual TLS Authentication

In the preceding section, you learned how clients authenticate servers by using the server’s certificate and a trusted third-party certificate or by configuring the client to explicitly trust the server’s certificate. 

Servers can authenticate clients in the same manner. This is particularly useful in zerotrust network infrastructures, where clients and servers must each prove their identities.

For example, you may have a client outside your network that must present a certificate to a proxy before the proxy will allow the client to access your trusted network resources. Likewise, the client authenticates the certificate presented by your proxy to make sure it’s talking to your proxy and not one controlled by a malicious actor.

You can instruct your server to set up TLS sessions with only authenticated clients. Those clients would have to present a certificate signed by a trusted certificate authority or pinned to the server.

clients cannot use the certificates generated with `$GOROOT/src/crypto/tls/generate_cert.go` for client authentication. Instead, you need to create your own certificate and private key.

### Generating Certificates for Authentication

Go’s standard library contains everything you need to generate your own certificates using the elliptic curve digital signature algorithm (ECDSA) and the P-256 elliptic curve.

check out `cert.go`.
```bash
go run cert.go -cert serverCert.pem -key serverKey.pem -host localhost
go run cert.go -cert clientCert.pem -key clientKey.pem -host localhost
```
### Implementing Mutual TLS

Now that you’ve generated certificate and private-key pairs for both the server and the client, you can start writing their code. Let’s write a test that implements mutual TLS authentication between our echo server and a client.

check `caCertPool` in `cert.go`.

Both the client and server use the caCertPool function to create a new X.509 certificate pool. The function accepts the file path to a PEM-encoded certificate, which you read in and append to the new certificate pool.

The certificate pool serves as a source of trusted certificates. The client puts the server’s certificate in its certificate pool, and vice versa.

check `TestMutualTLSAuthentication` in `mutual_test.go`.


Remember that in `cert.go`, you defined the IPAddresses and DNSNames slices of the template used to generate your client’s certificate. These values populate the common name and alternative names portions of the client’s certificate. You learned that Go’s TLS client uses these values to authenticate the server. But *the server does not use these values from the client’s certificate to authenticate the client*.

Since you’re implementing mutual TLS authentication, you need to make some changes to the server’s certificate verification process so that it authenticates the client’s IP address or hostnames against the client certificate’s common name and alternative names.

To do that, the server at the very least needs to know the client’s IP address. The only way you can get client connection information before certificate verification is by defining the tls.Config’s `GetConfigForClient` method.

This method allows you to define a function that receives the *tls.ClientHelloInfo object created as part of the TLS handshake process with the client. From this, you can retrieve the client’s IP address. But first, you need to return a proper TLS configuration.

Since you want to augment the usual certificate verification process on
the server, you define an appropriate function and assign it to the TLS configuration’s `VerifyPeerCertificate` method. The server calls this method
after the normal certificate verification checks. The only check you’re performing above and beyond the normal checks is to verify the client’s hostname with the leaf certificate’s common name and alternative names.

The leaf certificate is the last certificate in the certificate chain given to
the server by the client. The leaf certificate contains the client’s public key.
All other certificates in the chain are intermediate certificates used to verify
the authenticity of the leaf certificate and culminate with the certificate
authority’s certificate. You’ll find each leaf certificate at index 0 in each
verifiedChains slice. In other words, you can find the leaf certificate of the
first chain at verifiedChains[0][0]. If the server calls your function assigned
to the VerifyPeerCertificate method, the leaf certificate in the first chain
exists at a minimum