package main

import (
	"bytes"
	"log"
	"testing"
)

func TestLoggingConfiguration(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("Test log message")
	logOutput := buf.String()

	if len(logOutput) == 0 || !containsTimestamp(logOutput) {
		t.Errorf("Log output does not contain a timestamp: %s", logOutput)
	}
}

func containsTimestamp(logOutput string) bool {
	// Simple check for timestamp format (e.g., "2006/01/02 15:04:05.000000")
	return len(logOutput) > 20 && logOutput[4] == '/' && logOutput[7] == '/' && logOutput[10] == ' '
}
