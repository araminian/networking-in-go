package main

import (
	"io"
	"net"
)

func proxyConn(src, dst string) error {
	connSrc, err := net.Dial("tcp", src)
	if err != nil {
		return err
	}
	defer connSrc.Close()

	connDst, err := net.Dial("tcp", dst)
	if err != nil {
		return err
	}
	defer connDst.Close()

	// connDst replices to connSrc
	go func() {
		/*
			You don’t need to worry about leaking this goroutine, since io.Copy will
			return when either connection is closed.
		*/
		_, _ = io.Copy(connSrc, connDst)
	}()

	// connSrc replices to connDst
	_, err = io.Copy(connDst, connSrc)

	return err
}

/*
you could proxy data from a network connection to os.Stdout,
*bytes.Buffer, *os.File, or any number of objects that implement the io.Writer
interface
*/
func proxy(from io.Reader, to io.Writer) error {
	fromWriter, isFromWriter := from.(*net.TCPConn)
	toReader, isToReader := to.(*net.TCPConn)

	if isFromWriter && isToReader {
		go func() {
			_, _ = io.Copy(fromWriter, toReader)
		}()
	}

	_, err := io.Copy(to, from)

	return err
}
