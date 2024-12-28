package main

import (
	"context"
	"fmt"
	"io"
	"time"
)

const (
	defaultPingInterval = 30 * time.Second
)

// The Pinger function writes ping messages to a given writer at regular intervals.
// It also allows for resetting the interval to a new value.
// You create a buffered channel and put a duration on it to set the timer’s initial interval
func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration
	select {
	// if the context is done, return
	case <-ctx.Done():
		return
	// if the reset channel is set, use the interval from the reset channel
	case interval = <-reset:
	default:
	}

	if interval <= 0 {
		// if the interval is not set, use the default interval
		interval = defaultPingInterval
	}

	// create a timer that will send a message to the writer after the interval
	timer := time.NewTimer(interval)
	defer func() {
		// stop the timer and drain the channel if the timer is not stopped
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		select {
		// if the context is done, return
		case <-ctx.Done():
			return
		// if the reset channel is set, use the interval from the reset channel
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		// if the timer is fired, write a ping message to the writer
		case <-timer.C:
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}
		}
		timer.Reset(interval)
	}
}

func ExamplePinger() {

	ctx, cancel := context.WithCancel(context.Background())

	r, w := io.Pipe() // in memory pipe

	done := make(chan struct{})
	resetTimer := make(chan time.Duration, 1)
	resetTimer <- 1 * time.Second

	// start the pinger
	go func() {
		Pinger(ctx, w, resetTimer) // it's blocking until the context is done
		close(done)                // notify the main goroutine that the pinger has finished
	}()

	receivePing := func(d time.Duration, r io.Reader) {
		if d > 0 {
			fmt.Printf("resetting timer to %s\n", d)
			resetTimer <- d
		}

		now := time.Now()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			fmt.Printf("read error: %v\n", err)
		}
		fmt.Printf("read %q (%s)\n", buf[:n], time.Since(now).Round(100*time.Millisecond))
	}

	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d:\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}

	cancel() // cancel the context, which will stop the pinger
	<-done   // wait for the pinger to finish

}
