package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
)

func TestOutputTable(t *testing.T) {
	report := createTestReport()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTable(report)
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTable failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Check for expected content
	if !strings.Contains(output, "Coverage Report") {
		t.Error("Table output should contain 'Coverage Report'")
	}

	if !strings.Contains(output, "test1.go") {
		t.Error("Table should contain test1.go")
	}

	if !strings.Contains(output, "test2.go") {
		t.Error("Table should contain test2.go")
	}

	if !strings.Contains(output, "OVERALL") {
		t.Error("Table should contain OVERALL summary")
	}
}

func TestOutputJSON(t *testing.T) {
	report := createTestReport()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(report)
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputJSON failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("JSON output is not valid: %v", err)
	}

	// Check content
	if !strings.Contains(output, "test1.go") {
		t.Error("JSON should contain test1.go")
	}
}

func TestOutputCSV(t *testing.T) {
	report := createTestReport()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputCSV(report)
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputCSV failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Parse as CSV
	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV output is not valid: %v", err)
	}

	// Check header
	if len(records) < 1 {
		t.Fatal("CSV should have at least a header row")
	}

	header := records[0]
	expectedHeaders := []string{"File", "Coverage %", "Covered Lines", "Total Lines"}
	for i, expected := range expectedHeaders {
		if i >= len(header) || header[i] != expected {
			t.Errorf("Expected header[%d] to be '%s', got '%s'", i, expected, header[i])
		}
	}

	// Check we have data rows
	if len(records) < 2 {
		t.Error("CSV should have data rows")
	}
}

func TestFilterBelowThreshold(t *testing.T) {
	report := createTestReport()

	// Filter for files below 80%
	filtered := filterBelowThreshold(report, 80.0)

	// Should only have test1.go (75%)
	if len(filtered.Files) != 1 {
		t.Errorf("Expected 1 file below 80%%, got %d", len(filtered.Files))
	}

	if _, exists := filtered.Files["test1.go"]; !exists {
		t.Error("Filtered report should contain test1.go")
	}

	if _, exists := filtered.Files["test2.go"]; exists {
		t.Error("Filtered report should not contain test2.go (100%)")
	}
}

func TestFilterBelowThresholdNoMatches(t *testing.T) {
	report := createTestReport()

	// Filter for files below 50% - should match nothing
	filtered := filterBelowThreshold(report, 50.0)

	if len(filtered.Files) != 0 {
		t.Errorf("Expected 0 files below 50%%, got %d", len(filtered.Files))
	}
}

func TestOutputTableEmptyReport(t *testing.T) {
	report := models.NewCoverageReport()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTable(report)
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTable with empty report failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "No files found") {
		t.Error("Empty report should indicate no files found")
	}
}

// Helper function to create a test report
func createTestReport() *models.CoverageReport {
	report := models.NewCoverageReport()
	report.TestName = "test-suite"

	// Add file 1 with 75% coverage
	file1 := &models.FileCoverage{
		FileName:     "test1.go",
		TotalLines:   4,
		CoveredLines: 3,
		CoveragePct:  75.0,
		Lines: map[int]models.LineCoverage{
			1: {LineNumber: 1, ExecutionCount: 1},
			2: {LineNumber: 2, ExecutionCount: 1},
			3: {LineNumber: 3, ExecutionCount: 1},
			4: {LineNumber: 4, ExecutionCount: 0},
		},
	}
	report.AddFile(file1)

	// Add file 2 with 100% coverage
	file2 := &models.FileCoverage{
		FileName:     "test2.go",
		TotalLines:   2,
		CoveredLines: 2,
		CoveragePct:  100.0,
		Lines: map[int]models.LineCoverage{
			1: {LineNumber: 1, ExecutionCount: 5},
			2: {LineNumber: 2, ExecutionCount: 3},
		},
	}
	report.AddFile(file2)

	return report
}
