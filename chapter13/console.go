package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Example_zapConsole() {

	/*
		The console encoder uses tabs to separate fields. It takes instruction
		from your encoder configuration to determine which fields to include and
		how to format each.
	*/
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zapcore.InfoLevel,
		),
	)
	defer func() { _ = zl.Sync() }()
	console := zl.Named("[console]")
	console.Info("this is logged by the logger")
	console.Debug("this is below the logger's threshold and won't log")
	console.Error("this is also logged by the logger")
	// Output:
	// info [console] this is logged by the logger
	// error [console] this is also logged by the logger
}
