package main

import (
	"os"
	"strings"
	"testing"
)

// TestRunParseWithGoCoverage tests parsing a real Go coverage file
func TestRunParseWithGoCoverage(t *testing.T) {
	// Create temp Go coverage file
	tmpFile, err := os.CreateTemp("", "coverage-*.out")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Write valid Go coverage data
	goCoverage := `mode: set
github.com/Chapati-Systems/covpeek/pkg/models/coverage.go:12.49,14.2 1 1
github.com/Chapati-Systems/covpeek/pkg/models/coverage.go:17.56,19.2 1 1
`
	_, _ = tmpFile.WriteString(goCoverage)
	_ = tmpFile.Close()

	// Reset flags
	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})

	// We don't capture output since fmt.Printf writes directly to os.Stdout
	// This test just verifies the code path executes without error
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseWithJSONOutput tests JSON output format
func TestRunParseWithJSONOutput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	lcovData := `TN:test
SF:file.go
DA:1,1
DA:2,0
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	// Reset flags
	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--output", "json"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseWithCSVOutput tests CSV output format
func TestRunParseWithCSVOutput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	lcovData := `TN:test
SF:file.go
DA:1,1
DA:2,0
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	// Reset flags
	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--output", "csv"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseWithBelowThreshold tests the --below flag
func TestRunParseWithBelowThreshold(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	lcovData := `TN:test
SF:file1.go
DA:1,1
DA:2,0
end_of_record
SF:file2.go
DA:1,1
DA:2,1
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	// Reset flags
	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--below", "80"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseWithForceFormat tests forcing format detection
func TestRunParseWithForceFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	lcovData := `TN:test
SF:file.go
DA:1,1
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	// Reset flags
	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--format", "lcov"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestSimple is a minimal test to verify runParse executes
func TestSimple(t *testing.T) {
	// Use the existing sample.lcov file from testdata
	coverageFile = "../../testdata/sample.lcov"
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", "../../testdata/sample.lcov"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestRunParseHelpArgument tests that help argument shows help
func TestRunParseHelpArgument(t *testing.T) {
	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"help"})
	err := rootCmd.Execute()
	// Help should not error
	if err != nil {
		t.Fatalf("Expected no error for help, got: %v", err)
	}
}

// TestBadgeGeneration tests the badge command with real coverage file
func TestBadgeGeneration(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "badge-*.svg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Use the existing sample.lcov
	coverageFile = "../../testdata/sample.lcov"
	badgeFile = "../../testdata/sample.lcov"
	badgeOutput = tmpFile.Name()
	badgeLabel = "coverage"
	badgeStyle = "flat"

	rootCmd.SetArgs([]string{"badge", "--file", "../../testdata/sample.lcov", "--output", tmpFile.Name()})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check file exists and has SVG content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Error("SVG file is empty")
	}
	if !strings.Contains(string(content), "<svg") {
		t.Error("File does not contain SVG")
	}
	if !strings.Contains(string(content), "coverage") {
		t.Error("SVG does not contain label")
	}
}
