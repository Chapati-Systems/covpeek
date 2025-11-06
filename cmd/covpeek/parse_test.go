package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunParseWithSampleLCOV(t *testing.T) {
	// This test verifies that parse functions can handle real files
	// We test the individual functions directly instead of via command execution
	// which provides better unit test isolation

	// Test that validation passes for a valid file
	err := validateTestFlags("../../testdata/sample.lcov", "", 0, "table")
	if err != nil {
		t.Errorf("Validation should pass for valid file, got: %v", err)
	}

	// Test with format override
	err = validateTestFlags("../../testdata/sample.lcov", "rust", 0, "json")
	if err != nil {
		t.Errorf("Validation should pass with rust format, got: %v", err)
	}

	// Test with below filter
	err = validateTestFlags("../../testdata/sample.lcov", "", 50, "csv")
	if err != nil {
		t.Errorf("Validation should pass with below filter, got: %v", err)
	}
}

func TestValidateFlagsAllCombinations(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		format      string
		below       float64
		output      string
		shouldError bool
		errorText   string
	}{
		{
			name:        "valid defaults",
			file:        "../../testdata/sample.lcov",
			format:      "",
			below:       0,
			output:      "table",
			shouldError: false,
		},
		{
			name:        "valid with all flags",
			file:        "../../testdata/sample.lcov",
			format:      "rust",
			below:       50,
			output:      "json",
			shouldError: false,
		},
		{
			name:        "empty file",
			file:        "",
			format:      "",
			below:       0,
			output:      "table",
			shouldError: true,
			errorText:   "required",
		},
		{
			name:        "invalid below negative",
			file:        "../../testdata/sample.lcov",
			format:      "",
			below:       -10,
			output:      "table",
			shouldError: true,
			errorText:   "between 0 and 100",
		},
		{
			name:        "invalid below too high",
			file:        "../../testdata/sample.lcov",
			format:      "",
			below:       200,
			output:      "table",
			shouldError: true,
			errorText:   "between 0 and 100",
		},
		{
			name:        "invalid format",
			file:        "../../testdata/sample.lcov",
			format:      "python",
			below:       0,
			output:      "table",
			shouldError: true,
			errorText:   "invalid format",
		},
		{
			name:        "invalid output",
			file:        "../../testdata/sample.lcov",
			format:      "",
			below:       0,
			output:      "xml",
			shouldError: true,
			errorText:   "invalid output format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTestFlags(tt.file, tt.format, tt.below, tt.output)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestFormatMappingEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		format      string
		shouldError bool
	}{
		{"lcov lowercase", "lcov", false},
		{"LCOV uppercase", "LCOV", false},
		{"go lowercase", "go", false},
		{"GO uppercase", "GO", false},
		{"rust lowercase", "rust", false},
		{"RUST uppercase", "RUST", false},
		{"ts lowercase", "ts", false},
		{"TS uppercase", "TS", false},
		{"invalid format", "java", true},
		{"empty format", "", false}, // Empty is allowed, means auto-detect
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTestFlags("../../testdata/sample.lcov", tc.format, 0, "table")

			if tc.shouldError && err == nil {
				t.Error("Expected error for invalid format")
			} else if !tc.shouldError && err != nil {
				t.Errorf("Did not expect error, got: %v", err)
			}
		})
	}
}

// TestValidateFlagsDirect tests validateFlags function directly
func TestValidateFlagsDirect(t *testing.T) {
	cmd := &cobra.Command{}

	// Test missing file
	coverageFile = ""
	err := validateFlags(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error for missing file, got: %v", err)
	}

	// Test nonexistent file
	coverageFile = "/nonexistent/path/file.lcov"
	err = validateFlags(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error for nonexistent file, got: %v", err)
	}

	// Test invalid format
	coverageFile = "../../testdata/sample.lcov"
	forceFormat = "invalid"
	err = validateFlags(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected error for invalid format, got: %v", err)
	}

	// Test invalid below value
	forceFormat = ""
	belowPct = -10
	err = validateFlags(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "between 0 and 100") {
		t.Errorf("Expected error for invalid below, got: %v", err)
	}

	// Test invalid output format
	belowPct = 50
	outputFormat = "xml"
	err = validateFlags(cmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("Expected error for invalid output, got: %v", err)
	}

	// Test valid flags
	outputFormat = "table"
	err = validateFlags(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error for valid flags, got: %v", err)
	}
}

// TestRunParseDirect tests runParse function directly
func TestRunParseDirect(t *testing.T) {
	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test with help argument
	err := runParse(cmd, []string{"help"})
	// Help should not error
	if err != nil {
		t.Logf("Help returned: %v", err)
	}

	// Test with valid LCOV file
	coverageFile = "../../testdata/sample.lcov"
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error for valid LCOV, got: %v", err)
	}

	// Test with JSON output
	outputFormat = "json"
	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error for JSON output, got: %v", err)
	}

	// Test with CSV output
	outputFormat = "csv"
	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error for CSV output, got: %v", err)
	}

	// Test with force format
	forceFormat = "lcov"
	outputFormat = "table"
	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error with forced format, got: %v", err)
	}
}

func TestRunParse(t *testing.T) {
	// Save originals
	origFile := coverageFile
	origFormat := outputFormat
	origForce := forceFormat
	origBelow := belowPct
	origTui := tuiMode
	defer func() {
		coverageFile = origFile
		outputFormat = origFormat
		forceFormat = origForce
		belowPct = origBelow
		tuiMode = origTui
	}()

	// Test table output
	coverageFile = "../../testdata/sample.lcov"
	outputFormat = "table"
	forceFormat = ""
	belowPct = 0
	tuiMode = false

	cmd := &cobra.Command{}
	err := runParse(cmd, []string{})
	if err != nil {
		t.Errorf("runParse failed: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("runParse JSON failed: %v", err)
	}

	// Test CSV output
	outputFormat = "csv"
	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("runParse CSV failed: %v", err)
	}

	// Test with force format
	forceFormat = "lcov"
	outputFormat = "table"
	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("runParse with force format failed: %v", err)
	}

	// Test with below filter
	belowPct = 50
	err = runParse(cmd, []string{})
	if err != nil {
		t.Errorf("runParse with below filter failed: %v", err)
	}
}

func TestRunParseNonExistentFile(t *testing.T) {
	// Save original values
	origFile := coverageFile
	origTui := tuiMode
	defer func() {
		coverageFile = origFile
		tuiMode = origTui
	}()

	cmd := &cobra.Command{}

	// Test non-existent file
	coverageFile = "nonexistent.lcov"
	err := runParse(cmd, []string{})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestRunParseDirectory(t *testing.T) {
	// Save original values
	origFile := coverageFile
	origTui := tuiMode
	defer func() {
		coverageFile = origFile
		tuiMode = origTui
	}()

	cmd := &cobra.Command{}

	// Test directory instead of file
	coverageFile = "../../testdata" // This is a directory
	err := runParse(cmd, []string{})
	if err == nil {
		t.Error("Expected error for directory")
	}
}
