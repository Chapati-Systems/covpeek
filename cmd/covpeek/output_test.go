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
	"github.com/spf13/cobra"
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
	if !strings.Contains(output, "File") {
		t.Error("Table output should contain 'File' header")
	}

	if !strings.Contains(output, "Total Lines") {
		t.Error("Table output should contain 'Total Lines' header")
	}

	if !strings.Contains(output, "test1.go") {
		t.Error("Table should contain test1.go")
	}

	if !strings.Contains(output, "test2.go") {
		t.Error("Table should contain test2.go")
	}

	if !strings.Contains(output, "Overall") {
		t.Error("Table should contain Overall summary")
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

func TestRunParseHelp(t *testing.T) {
	// Save original values
	origFile := coverageFile
	origTui := tuiMode
	defer func() {
		coverageFile = origFile
		tuiMode = origTui
	}()

	cmd := &cobra.Command{}

	// Test help argument
	err := runParse(cmd, []string{"help"})
	// This should not error, but we can't easily test the help output
	// Just test that it doesn't crash
	_ = err // We expect this might return an error or nil depending on implementation
}

func TestFilterBelowThresholdNoMatches(t *testing.T) {
	report := createTestReport()

	// Filter for files below 50% - should match nothing
	filtered := filterBelowThreshold(report, 50.0)

	if len(filtered.Files) != 0 {
		t.Errorf("Expected 0 files below 50%%, got %d", len(filtered.Files))
	}
}

func TestOutputCSVEmptyReport(t *testing.T) {
	report := models.NewCoverageReport()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputCSV(report)
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputCSV with empty report failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should have header only
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line for empty report, got %d", len(lines))
	}

	if !strings.Contains(output, "File") {
		t.Error("CSV should contain header")
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

	// Should print "No files found in coverage report"
	if !strings.Contains(output, "No files found in coverage report") {
		t.Error("Table should contain 'No files found in coverage report'")
	}
}

func TestOutputJSONEmptyReport(t *testing.T) {
	report := models.NewCoverageReport()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(report)
	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputJSON with empty report failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("JSON output is not valid: %v", err)
	}

	// Should have empty files
	if files, ok := result["Files"].(map[string]interface{}); !ok || len(files) != 0 {
		t.Error("Expected empty Files map")
	}
}

func TestValidateFlags(t *testing.T) {
	// Save original values
	origFile := coverageFile
	origBelow := belowPct
	origFormat := outputFormat
	origForce := forceFormat
	defer func() {
		coverageFile = origFile
		belowPct = origBelow
		outputFormat = origFormat
		forceFormat = origForce
	}()

	cmd := &cobra.Command{}

	// Test missing file flag
	coverageFile = ""
	err := validateFlags(cmd, []string{})
	if err == nil {
		t.Error("Expected error for missing file flag")
	}

	// Test file doesn't exist
	coverageFile = "nonexistent.lcov"
	err = validateFlags(cmd, []string{})
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	// Test directory instead of file
	coverageFile = "../../testdata" // This is a directory
	err = validateFlags(cmd, []string{})
	if err == nil {
		t.Error("Expected error for directory instead of file")
	}

	// Test invalid below percentage
	coverageFile = "../../testdata/sample.lcov"
	belowPct = 150 // Invalid
	err = validateFlags(cmd, []string{})
	if err == nil {
		t.Error("Expected error for invalid below percentage")
	}

	belowPct = -10 // Invalid
	err = validateFlags(cmd, []string{})
	if err == nil {
		t.Error("Expected error for negative below percentage")
	}

	// Test invalid output format
	belowPct = 0
	outputFormat = "invalid"
	err = validateFlags(cmd, []string{})
	if err == nil {
		t.Error("Expected error for invalid output format")
	}

	// Test invalid force format
	outputFormat = "table"
	forceFormat = "invalid"
	err = validateFlags(cmd, []string{})
	if err == nil {
		t.Error("Expected error for invalid force format")
	}

	// Test valid case
	forceFormat = ""
	err = validateFlags(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error for valid flags, got: %v", err)
	}
}

func TestFilterBelowThreshold(t *testing.T) {
	report := createTestReport()

	// Test filtering
	filtered := filterBelowThreshold(report, 80.0)
	if len(filtered.Files) != 1 {
		t.Errorf("Expected 1 file below 80%%, got %d", len(filtered.Files))
	}

	if _, exists := filtered.Files["test1.go"]; !exists {
		t.Error("Filtered report should contain test1.go")
	}

	// Test no matches
	filtered = filterBelowThreshold(report, 50.0)
	if len(filtered.Files) != 0 {
		t.Errorf("Expected 0 files below 50%%, got %d", len(filtered.Files))
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

func TestTableModel(t *testing.T) {
	report := createTestReport()

	// Test that newTableModel works without panicking
	model := newTableModel(report)
	if len(model.table.Rows()) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(model.table.Rows()))
	}

	// Test sorting functionality
	model.sortByColumn(0) // Sort by file name (descending initially)
	rows := model.table.Rows()
	if rows[0][0] != "test2.go" { // test2.go comes before test1.go in descending alphabetical order
		t.Errorf("Expected first row to be test2.go after sorting by name, got %s", rows[0][0])
	}

	model.sortByColumn(3) // Sort by coverage (descending)
	rows = model.table.Rows()
	if rows[0][3] != "100.00" {
		t.Errorf("Expected first row to have 100.00%% coverage after sorting, got %s", rows[0][3])
	}

	// Test View method doesn't panic
	view := model.View()
	if !strings.Contains(view, "Coverage Report") {
		t.Error("View should contain title")
	}
	if !strings.Contains(view, "File") {
		t.Error("View should contain File header")
	}
	if !strings.Contains(view, "Total Lines") {
		t.Error("View should contain Total Lines header")
	}

	// Test Init
	cmd := model.Init()
	_ = cmd // Just check it doesn't panic
}
