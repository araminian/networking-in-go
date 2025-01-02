package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func Example_logLevels() {
	lDebug := log.New(os.Stdout, "DEBUG: ", log.Lshortfile)
	logFile := new(bytes.Buffer)
	w := SustainedMultiWriter(logFile, lDebug.Writer())
	lError := log.New(w, "ERROR: ", log.Lshortfile)
	fmt.Println("standard output:")
	lError.Print("cannot communicate with the database")
	lDebug.Print("you cannot hum while holding your nose")
	fmt.Print("\nlog file contents:\n", logFile.String())
	// Output:
	// standard output:
	// ERROR: log_test.go:43: cannot communicate with the database
	// DEBUG: log_test.go:44: you cannot hum while holding your nose
	//
	// log file contents:
	// ERROR: log_test.go:43: cannot communicate with the database
}
