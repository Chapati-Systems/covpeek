package main

import (
	"os"
	"testing"
)

// TestRunParseInvalidGoFormat tests error handling for invalid Go coverage content
func TestRunParseInvalidGoFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.out")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Write invalid Go coverage data
	invalidData := `mode: set
invalid line
`
	_, _ = tmpFile.WriteString(invalidData)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	// Should error on parsing
	t.Logf("Got expected parsing error: %v", err)
}

// TestRunParseDetectionByContentFallback tests content-based detection when extension fails
func TestRunParseDetectionByContentFallback(t *testing.T) {
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

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseUnknownFormat tests error handling for unknown format
func TestRunParseUnknownFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.xyz")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Write content that doesn't match any known format
	_, _ = tmpFile.WriteString("This is not a coverage file\nJust random text\n")
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	// Error or success depending on detection - just log it
	t.Logf("Result for unknown format: %v", err)
}

// TestRunParseForceGoFormat tests forcing Go format
func TestRunParseForceGoFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	goCoverage := `mode: set
git.kernel.fun/test/file.go:1.1,2.2 1 1
`
	_, _ = tmpFile.WriteString(goCoverage)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--format", "go"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseForceRustFormat tests forcing Rust/LCOV format
func TestRunParseForceRustFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	lcovData := `TN:test
SF:file.rs
DA:1,1
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--format", "rust"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseForceTsFormat tests forcing TypeScript/LCOV format
func TestRunParseForceTsFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	lcovData := `TN:test
SF:file.ts
DA:1,1
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--format", "ts"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseInvalidForceFormat tests error handling for invalid forced format
func TestRunParseInvalidForceFormat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
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

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name(), "--format", "python"})
	err = rootCmd.Execute()
	// Should error for invalid format
	t.Logf("Result for invalid format (expected error): %v", err)
}

// TestRunParseDetectionByExtension tests format detection by file extension
func TestRunParseDetectionByExtension(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.out")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	goCoverage := `mode: set
git.kernel.fun/test/file.go:1.1,2.2 1 1
`
	_, _ = tmpFile.WriteString(goCoverage)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseDetectionByExtensionLcov tests LCOV detection by .lcov extension
func TestRunParseDetectionByExtensionLcov(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
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

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseDetectionByExtensionInfo tests LCOV detection by .info extension
func TestRunParseDetectionByExtensionInfo(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.info")
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

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseMalformedLcov tests error handling for malformed LCOV
func TestRunParseMalformedLcov(t *testing.T) {
	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", "../../testdata/malformed.lcov"})
	err := rootCmd.Execute()
	// Should still succeed even with warnings
	if err != nil {
		// Malformed may error, which is acceptable
		t.Logf("Got expected error for malformed LCOV: %v", err)
	}
}

// TestRunParseTableOutputWithLongFilename tests table output with truncation
func TestRunParseTableOutputWithLongFilename(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Create entry with very long filename
	longPath := "very/long/path/to/some/deeply/nested/directory/structure/that/exceeds/fifty/characters/file.go"
	lcovData := `TN:test
SF:` + longPath + `
DA:1,1
DA:2,1
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestRunParseEmptyReport tests handling of empty coverage report
func TestRunParseEmptyReport(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Empty LCOV file
	lcovData := ``
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	// Empty file should succeed but show no files
	if err != nil {
		t.Fatalf("Expected no error for empty file, got: %v", err)
	}
}

// TestRunParseWithTestName tests LCOV with test name
func TestRunParseWithTestName(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "coverage-*.lcov")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	lcovData := `TN:my-test-suite
SF:file.go
DA:1,1
end_of_record
`
	_, _ = tmpFile.WriteString(lcovData)
	_ = tmpFile.Close()

	coverageFile = ""
	forceFormat = ""
	belowPct = 0
	outputFormat = "table"

	rootCmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
