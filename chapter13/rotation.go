package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestZapLogRotation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),

			/*
				*lumberjack.Logger does not
				implement the zapcore.WriteSyncer. It, too, lacks a Sync method. Therefore,
				you need to wrap it in a call to zapcore.AddSync
			*/
			zapcore.AddSync(
				&lumberjack.Logger{
					Filename:   filepath.Join(tempDir, "debug.log"),
					Compress:   true,
					LocalTime:  true,
					MaxAge:     7,
					MaxBackups: 5,
					MaxSize:    100,
				},
			),
			zapcore.DebugLevel,
		),
	)
	defer func() { _ = zl.Sync() }()
	zl.Debug("debug message written to the log file")
}
