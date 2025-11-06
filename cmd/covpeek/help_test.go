package main

import (
	"bytes"
	"testing"
)

func TestHelpCommand(t *testing.T) {
	// Reset rootCmd for testing
	rootCmd.SetArgs([]string{"help"})

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Execute
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("help command produced no output")
	}

	// Check for expected content
	if !bytes.Contains(buf.Bytes(), []byte("covpeek")) {
		t.Errorf("help output doesn't contain 'covpeek': %s", output)
	}

	t.Logf("Help output length: %d bytes", len(output))
	t.Logf("Help output:\n%s", output)
}

func TestHelpFlag(t *testing.T) {
	// Reset rootCmd for testing
	rootCmd.SetArgs([]string{"--help"})

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Execute (should not return error for help)
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("--help flag failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Fatal("--help flag produced no output")
	}

	t.Logf("Help flag output length: %d bytes", len(output))
}
