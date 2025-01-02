package main

import (
	"bytes"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Example_zapInfoFileDebugConsole() {
	/*
		You’re using *bytes.Buffer to act as a mock log file. The only problem
		with this is that *bytes.Buffer does not have a Sync method and does
		not implement the zapcore.WriteSyncer interface. Thankfully, Zap includes a helper
		function named zapcore.AddSync that intelligently adds a no-op Sync method
		to an io.Writer.
	*/
	logFile := new(bytes.Buffer)
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(zapcore.AddSync(logFile)),
			zapcore.InfoLevel,
		),
	)
	defer func() { _ = zl.Sync() }()
	zl.Debug("this is below the logger's threshold and won't log")
	zl.Error("this is logged by the logger")

	/*
		let’s experiment
		with Zap and create a new logger that can simultaneously
		write JSON log entries to a log file and console log entries to standard output.
	*/

	/*
		Zap’s WithOptions method clones the existing logger and configures
		the clone with the given options. You can use the zap.WrapCore function
		to modify the underlying zap.Core of the cloned logger. To mix things up,
		you make a copy of the encoder configuration and tweak it to instruct the
		encoder to output the level using all capital letters. Lastly, you use the
		zapcore.NewTee function, which is like the io.MultiWriter function, to return
		a zap.Core that writes to multiple cores. In this example, you’re passing
		in the existing core and a new core that writes debug-level log entries to
		standard output.
	*/
	zl = zl.WithOptions(
		zap.WrapCore(
			func(c zapcore.Core) zapcore.Core {
				ucEncoderCfg := encoderCfg
				ucEncoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
				return zapcore.NewTee(
					c,
					zapcore.NewCore(
						zapcore.NewConsoleEncoder(ucEncoderCfg),
						zapcore.Lock(os.Stdout),
						zapcore.DebugLevel,
					),
				)
			},
		),
	)

	/*
		When you use the cloned logger, both the log file and standard output
		receive any log entry at the info level or above, whereas only
		standard output receives debugging log entries
	*/
	fmt.Println("standard output:")
	zl.Debug("this is only logged as console encoding")
	zl.Info("this is logged as console encoding and JSON")
	fmt.Print("\nlog file contents:\n", logFile.String())
	// Output:
	// standard output:
	// DEBUG this is only logged as console encoding
	// INFO this is logged as console encoding and JSON
	//
	// log file contents:
	// {"level":"error","msg":"this is logged by the logger"}
	// {"level":"info","msg":"this is logged as console encoding and JSON"}
}
