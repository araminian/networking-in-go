package main

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Example_zapSampling() {
	/*
		The NewSamplerWithOptions function wraps zap.Core with sampling
		functionality. It requires three additional arguments: a sampling interval,
		the number of initial duplicate log entries to record, and an integer
		representing the nth duplicate log entry to record after that point. In this
		example, you are logging the first log entry, and then every third duplicate
		log entry that the logger receives in a one-second interval. Once the interval
		elapses, the logger starts over and logs the first entry, then every third duplicate
		for the remainder of the one-second interval.
	*/
	zl := zap.New(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderCfg),
				zapcore.Lock(os.Stdout),
				zapcore.DebugLevel,
			),
			time.Second, 1, 3,
		),
	)
	defer func() { _ = zl.Sync() }()

	for i := 0; i < 10; i++ {
		if i == 5 {
			time.Sleep(time.Second)
		}
		zl.Debug(fmt.Sprintf("%d", i))
		zl.Debug("debug message")
	}
	// Output:
	// {"level":"debug","msg":"0"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"1"}
	// {"level":"debug","msg":"2"}
	// {"level":"debug","msg":"3"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"4"}
	// {"level":"debug","msg":"5"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"6"}
	// {"level":"debug","msg":"7"}
	// {"level":"debug","msg":"8"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"9"}
}
