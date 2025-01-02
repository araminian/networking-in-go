package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net-c12/homework/v1"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var addr, caCertFn string

func init() {
	flag.StringVar(&addr, "address", "localhost:34443", "server address")
	flag.StringVar(&caCertFn, "ca-cert", "cert.pem", "CA certificate")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [flags] [add chore, ...|complete #]
		add add comma-separated chores
		complete complete designated chore
		Flags:
		`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func list(ctx context.Context, client homework.RobotMaidClient) error {
	chores, err := client.List(ctx, new(homework.Empty))
	if err != nil {
		return err
	}
	if len(chores.Chores) == 0 {
		fmt.Println("You have nothing to do!")
		return nil
	}
	fmt.Println("#\t[X]\tDescription")
	for i, chore := range chores.Chores {
		c := " "
		if chore.Complete {
			c = "X"
		}
		fmt.Printf("%d\t[%s]\t%s\n", i+1, c, chore.Description)
	}
	return nil
}

func add(ctx context.Context, client homework.RobotMaidClient,
	s string) error {
	chores := new(homework.Chores)
	for _, chore := range strings.Split(s, ",") {
		if desc := strings.TrimSpace(chore); desc != "" {
			chores.Chores = append(chores.Chores, &homework.Chore{
				Description: desc,
			})
		}
	}
	var err error
	if len(chores.Chores) > 0 {
		_, err = client.Add(ctx, chores)
	}
	return err
}

func complete(ctx context.Context, client homework.RobotMaidClient,
	s string) error {
	i, err := strconv.Atoi(s)
	if err == nil {
		_, err = client.Complete(ctx,
			&homework.CompleteRequest{ChoreNumber: int32(i)})
	}
	return err
}

func main() {
	flag.Parse()
	caCert, err := os.ReadFile(caCertFn)
	if err != nil {
		log.Fatal(err)
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal("failed to add certificate to pool")
	}
	/*
		But you want
		to pin the server’s certificate to the client connection. Therefore, you need
		to explicitly pass in a grpc.DialOption with the appropriate TLS credentials.
		This involves using the grpc.WithTransportCredentials function
	*/
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					CurvePreferences:         []tls.CurveID{tls.CurveP256},
					MinVersion:               tls.VersionTLS12,
					PreferServerCipherSuites: true,
					RootCAs:                  certPool,
					NextProtos:               []string{"h2"},
				},
			),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// call the server

	// instantiating a new gRPC client from the gRPC network connection
	rosie := homework.NewRobotMaidClient(conn)
	ctx := context.Background()
	switch strings.ToLower(flag.Arg(0)) {
	case "add":
		err = add(ctx, rosie, strings.Join(flag.Args()[1:], " "))
	case "complete":
		err = complete(ctx, rosie, flag.Arg(1))
	}
	if err != nil {
		log.Fatal(err)
	}
	err = list(ctx, rosie)
	if err != nil {
		log.Fatal(err)
	}
}
