package main

import (
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var encoderCfg = zapcore.EncoderConfig{

	/*
		key msg for the log message and the
		key name for the logger’s name in the log entry.
	*/
	MessageKey: "message",
	NameKey:    "name",

	/*
		the encoder
		configuration tells the encoder to use the key level for the logging level
		and encode the level name using all lowercase characters
	*/
	LevelKey:    "level",
	EncodeLevel: zapcore.LowercaseLevelEncoder,

	/*
		If the logger
		is configured to add caller details, you want the encoder to associate these
		details with the caller key and encode the details in an abbreviated format
	*/
	CallerKey:    "caller",
	EncodeCaller: zapcore.ShortCallerEncoder,

	/*
		you want the encoder to associate these
		details with the caller key and encode the details in an abbreviated format
	*/
	TimeKey:    "time",
	EncodeTime: zapcore.ISO8601TimeEncoder,
}

func Example_zapJSON() {

	/*
		The zap.New function accepts a zap.Core 1 and zero or more zap.Options.
		In this example, you’re passing the zap.AddCaller option, which instructs
		the logger to include the caller information in each log entry, and a field
		named version that inserts the runtime version in each log entry.
	*/
	zl := zap.New(
		zapcore.NewCore(
			/*
				The zap.Core consists of a JSON encoder using your encoder configuration
				, a zapcore.WriteSyncer , and the logging threshold
			*/
			zapcore.NewJSONEncoder(encoderCfg),
			/*
				If the zapcore.WriteSyncer isn’t safe for concurrent use,
				you can use zapcore.Lock to make it
				concurrency safe, as in this example.
			*/
			zapcore.Lock(os.Stdout),
			zapcore.DebugLevel,
		),
		zap.AddCaller(),
		zap.Fields(
			zap.String("version", runtime.Version()),
		),
	)

	/*
		Before you start using the logger, you want to make sure you defer a call
		to its Sync method to ensure all buffered data is written to the output.
	*/
	defer func() { _ = zl.Sync() }()

	/*
		The Zap logger includes seven log levels, in increasing severity: DebugLevel,
		InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, and FatalLevel. The
		InfoLevel is the default. DPanicLevel and PanicLevel entries will cause Zap to log
		the entry and then panic. An entry logged at the FatalLevel will cause Zap to
		call os.Exit(1) after writing the log entry. Since your logger is using DebugLevel,
		it will log all entries.

		I recommend you restrict the use of DPanicLevel and PanicLevel to
		development and FatalLevel to production, and only then for catastrophic
		startup errors, such as a failure to connect to the database.
	*/
	/*
		You can also assign the logger a name by calling its Named method
		and using the returned logger. By default, a logger has no name. A named
		logger will include a name key in the log entry, provided you defined one in
		the encoder configuration.
	*/
	example := zl.Named("example")
	example.Debug("test debug message")
	example.Info("test info message")
	// Output:
	// {"level":"debug","name":"example","caller":"ch13/zap_test.go:49",
	// "msg":"test debug message","version":"ago1.15.5"}
	// {"level":"info","name":"example","caller":"ch13/zap_test.go:50",
	// "msg":"test info message","version":"go1.15.5"}
}
