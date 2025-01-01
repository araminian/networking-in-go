package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

var (
	host = flag.String("host", "localhost",
		"Certificate's comma-separated host names and IPs")
	certFn = flag.String("cert", "clientcert.pem", "certificate file name")
	keyFn  = flag.String("key", "clientkey.pem", "private key file name")
	org    = flag.String("org", "cloudarmin", "organization name")
)

func main() {
	flag.Parse()

	/*
		The process of generating a certificate and a private key involves
		building a template in your code that you then encode to the X.509 format. Each certificate needs a serial number, which a certificate authority
		typically assigns. Since you’re generating your own self-signed certificate,
		you generate your own serial number using a cryptographically random,
		unsigned 128-bit integer
	*/
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1),
		128))
	if err != nil {
		log.Fatal(err)
	}

	notBefore := time.Now()

	/*

		You then create an x509.Certificate object that
		represents an X.509-formatted certificate and set various values, such as
		the serial number, the certificate’s subject, the validity lifetime, and various
		usages for this certificate. Since you want to use this certificate for client
		authentication, you must include the x509.ExtKeyUsageClientAuth value 2.
		If you omit this value, the server won’t be able to verify the certificate when
		presented by the client.
	*/
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{*org},
		},
		NotBefore: notBefore,
		NotAfter:  notBefore.Add(10 * 356 * 24 * time.Hour),
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// add the host names and IPs to the certificate template
	for _, h := range strings.Split(*host, ",") {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	/*
		You then generate a private key using the elliptic.P256() function, which
		creates a new elliptic curve key using the P-256 elliptic curve.
	*/
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	/*
		At this point, you have everything you need to generate the certificate. The x509.CreateCertificate function 5 accepts a source of entropy
		(crypto/rand’s Reader is ideal), the template for the new certificate, a parent
		certificate, a public key, and a corresponding private key.
	*/
	der, err := x509.CreateCertificate(rand.Reader, &template,
		&template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatal(err)
	}
	cert, err := os.Create(*certFn)
	if err != nil {
		log.Fatal(err)
	}

	/*
		It then returns a
		slice of bytes containing the Distinguished Encoding Rules (DER)–encoded
		certificate. You use your template for the parent certificate since the resulting certificate signs itself. All that’s left to do is create a new file, generate
		a new pem.Block with the DER-encoded byte slice, and PEM-encode everything to the new file
	*/
	err = pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err != nil {
		log.Fatal(err)
	}
	if err := cert.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", *certFn)

	key, err := os.OpenFile(*keyFn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600)
	if err != nil {
		log.Fatal(err)
	}

	/*
		We marshal the private key into a
		byte slice and, similarly, assign it to a new pem.Block before writing the
		PEM-encoded output to the private-key file
	*/
	privKey, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}
	err = pem.Encode(key, &pem.Block{Type: "EC PRIVATE KEY",
		Bytes: privKey})
	if err != nil {
		log.Fatal(err)
	}
	if err := key.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote", *keyFn)
}

// caCertPool is a helper function that reads a CA certificate file and returns a x509.CertPool.
func caCertPool(caCertFn string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caCertFn)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("failed to add certificate to pool")
	}
	return certPool, nil
}
