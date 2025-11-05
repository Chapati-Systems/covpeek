package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpOutput(t *testing.T) {
	// Reset command for testing
	rootCmd.SetArgs([]string{"--help"})
	
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	output := buf.String()
	
	// Verify help contains key elements
	expectedStrings := []string{
		"covpeek",
		"--file",
		"--format",
		"--below",
		"--output",
		"table",
		"json",
		"csv",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing expected string: %s", expected)
		}
	}
}

func TestHelpShowsExamples(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	output := buf.String()
	
	if !strings.Contains(output, "Examples:") {
		t.Error("Help output should contain Examples section")
	}
	
	if !strings.Contains(output, "coverage.lcov") {
		t.Error("Help examples should show coverage file examples")
	}
}
