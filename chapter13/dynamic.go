package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Example_zapDynamicDebugging() {
	/*
		Your code will watch for the level.debug file in the temporary directory.
		When the file is present, you’ll dynamically change the logger’s level to debug.
	*/
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	/*
		To do that, you need a new atomic leveler. By default, the atomic leveler
		uses the info level, which suits this example just fine. You pass in the atomic
		leveler when creating the core as opposed to specifying a log level itself.
	*/
	debugLevelFile := filepath.Join(tempDir, "level.debug")
	atomicLevel := zap.NewAtomicLevel()
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atomicLevel,
		),
	)
	defer func() { _ = zl.Sync() }()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()
	err = watcher.Add(tempDir)
	if err != nil {
		log.Fatal(err)
	}
	ready := make(chan struct{})
	go func() {
		defer close(ready)
		originalLevel := atomicLevel.Level()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name == debugLevelFile {
					switch {
					case event.Op&fsnotify.Create == fsnotify.Create:
						atomicLevel.SetLevel(zapcore.DebugLevel)
						ready <- struct{}{}
					case event.Op&fsnotify.Remove == fsnotify.Remove:
						atomicLevel.SetLevel(originalLevel)
						ready <- struct{}{}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				zl.Error(err.Error())
			}
		}
	}()

	zl.Debug("this is below the logger's threshold")
	df, err := os.Create(debugLevelFile)
	if err != nil {
		log.Fatal(err)
	}
	err = df.Close()
	if err != nil {
		log.Fatal(err)
	}
	<-ready
	zl.Debug("this is now at the logger's threshold")
	err = os.Remove(debugLevelFile)
	if err != nil {
		log.Fatal(err)
	}
	<-ready
	zl.Debug("this is below the logger's threshold again")
	zl.Info("this is at the logger's current threshold")
	// Output:
	// {"level":"debug","msg":"this is now at the logger's threshold"}
	// {"level":"info","msg":"this is at the logger's current threshold"}
}
