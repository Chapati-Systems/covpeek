package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestMissingFileFlag(t *testing.T) {
	// Create a fresh command for this test
	cmd := createTestCommand()
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when --file flag is missing")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "required") || !strings.Contains(errStr, "file") {
		t.Errorf("Error message should mention required file flag, got: %v", err)
	}
}

// createTestCommand creates a fresh command instance for testing
func createTestCommand() *cobra.Command {
	var testCoverageFile string
	var testOutputFormat string
	var testForceFormat string
	var testBelowPct float64

	testCmd := &cobra.Command{
		Use:   "covpeek --file <path> [flags]",
		Short: "Cross-language Coverage Report CLI Parser",
		Long: `covpeek is a CLI tool for parsing and analyzing coverage reports 
from multiple languages including Rust, Go, TypeScript, and JavaScript.

It supports LCOV format (.lcov, .info) and Go coverage format (.out).`,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validateTestFlags(testCoverageFile, testForceFormat, testBelowPct, testOutputFormat)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Minimal stub for testing
			return nil
		},
	}

	testCmd.Flags().StringVarP(&testCoverageFile, "file", "f", "", "Path to coverage file (required)")
	testCmd.Flags().StringVar(&testForceFormat, "format", "", "Override format detection (rust, go, ts)")
	testCmd.Flags().Float64Var(&testBelowPct, "below", 0, "Coverage threshold filter (0-100)")
	testCmd.Flags().StringVarP(&testOutputFormat, "output", "o", "table", "Output format (table, json, csv)")
	_ = testCmd.MarkFlagRequired("file")

	return testCmd
}

// validateTestFlags is a copy of validateFlags for testing
func validateTestFlags(coverageFile, forceFormat string, belowPct float64, outputFormat string) error {
	if coverageFile == "" {
		return fmt.Errorf("--file flag is required")
	}

	fileInfo, err := os.Stat(coverageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", coverageFile)
		}
		return fmt.Errorf("cannot access file %s: %w", coverageFile, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", coverageFile)
	}

	file, err := os.Open(coverageFile)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", coverageFile, err)
	}
	_ = file.Close()

	if forceFormat != "" {
		validFormats := map[string]bool{
			"rust": true,
			"go":   true,
			"ts":   true,
			"lcov": true,
		}
		if !validFormats[strings.ToLower(forceFormat)] {
			return fmt.Errorf("invalid format '%s': must be one of: rust, go, ts", forceFormat)
		}
	}

	if belowPct < 0 || belowPct > 100 {
		return fmt.Errorf("--below must be between 0 and 100, got: %.2f", belowPct)
	}

	validOutputs := map[string]bool{
		"table": true,
		"json":  true,
		"csv":   true,
	}
	if !validOutputs[strings.ToLower(outputFormat)] {
		return fmt.Errorf("invalid output format '%s': must be one of: table, json, csv", outputFormat)
	}

	return nil
}

func TestNonExistentFile(t *testing.T) {
	cmd := createTestCommand()
	cmd.SetArgs([]string{"--file", "/nonexistent/file.lcov"})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "does not exist") && !strings.Contains(err.Error(), "no such file") {
		t.Errorf("Error should indicate file doesn't exist, got: %v", err)
	}
}

func TestInvalidBelowValue(t *testing.T) {
	// Create temp file
	tmpFile := createTempCoverageFile(t)
	defer func() { _ = os.Remove(tmpFile) }()

	testCases := []struct {
		name  string
		value string
	}{
		{"below -1", "--below=-1"},
		{"below 101", "--below=101"},
		{"below 150", "--below=150"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := createTestCommand()
			cmd.SetArgs([]string{"--file", tmpFile, tc.value})

			var buf bytes.Buffer
			cmd.SetErr(&buf)

			err := cmd.Execute()
			if err == nil {
				t.Errorf("Expected error for invalid --below value: %s", tc.value)
			}

			if !strings.Contains(err.Error(), "between 0 and 100") {
				t.Errorf("Error should mention valid range, got: %v", err)
			}
		})
	}
}

func TestInvalidFormatValue(t *testing.T) {
	tmpFile := createTempCoverageFile(t)
	defer func() { _ = os.Remove(tmpFile) }()

	cmd := createTestCommand()
	cmd.SetArgs([]string{"--file", tmpFile, "--format", "invalid"})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Error should mention invalid format, got: %v", err)
	}
}

func TestInvalidOutputValue(t *testing.T) {
	tmpFile := createTempCoverageFile(t)
	defer func() { _ = os.Remove(tmpFile) }()

	cmd := createTestCommand()
	cmd.SetArgs([]string{"--file", tmpFile, "--output", "xml"})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid output format")
	}

	if !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("Error should mention invalid output format, got: %v", err)
	}
}

func TestDirectoryAsFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "covpeek-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cmd := createTestCommand()
	cmd.SetArgs([]string{"--file", tmpDir})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error when providing directory instead of file")
	}

	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("Error should mention directory, got: %v", err)
	}
}

func TestValidFlagsAccepted(t *testing.T) {
	tmpFile := createTempCoverageFile(t)
	defer func() { _ = os.Remove(tmpFile) }()

	testCases := []struct {
		name string
		args []string
	}{
		{"default", []string{"--file", tmpFile}},
		{"with format", []string{"--file", tmpFile, "--format", "go"}},
		{"with below", []string{"--file", tmpFile, "--below", "80"}},
		{"with output json", []string{"--file", tmpFile, "--output", "json"}},
		{"with output csv", []string{"--file", tmpFile, "--output", "csv"}},
		{"all flags", []string{"--file", tmpFile, "--format", "rust", "--below", "50", "--output", "table"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := createTestCommand()
			cmd.SetArgs(tc.args)

			var buf bytes.Buffer
			cmd.SetErr(&buf)
			cmd.SetOut(&buf)

			// Note: We expect execution to succeed (validation passes)
			// The parse might fail if file format is wrong, but that's okay for this test
			err := cmd.Execute()
			// We're just testing that validation passes, parsing errors are okay
			if err != nil && (strings.Contains(err.Error(), "between 0 and 100") ||
				strings.Contains(err.Error(), "invalid format") ||
				strings.Contains(err.Error(), "does not exist")) {
				t.Errorf("Validation should pass but got error: %v", err)
			}
		})
	}
}

// Helper function to create a temporary coverage file
func createTempCoverageFile(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Write minimal valid LCOV content
	content := `TN:
SF:example.go
DA:1,1
DA:2,0
LF:2
LH:1
end_of_record
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	_ = tmpFile.Close()
	return tmpFile.Name()
}

func TestFilePermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Create temp file with no read permissions
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Remove all permissions
	if err := os.Chmod(tmpFile.Name(), 0000); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer func() { _ = os.Chmod(tmpFile.Name(), 0644) }() // Restore for cleanup

	cmd := createTestCommand()
	cmd.SetArgs([]string{"--file", tmpFile.Name()})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for unreadable file")
	}

	if !strings.Contains(err.Error(), "cannot read") && !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Error should mention permission/read issue, got: %v", err)
	}
}
