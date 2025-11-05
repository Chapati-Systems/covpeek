package models

import (
	"testing"
)

func TestNewCoverageReport(t *testing.T) {
	report := NewCoverageReport()

	if report == nil {
		t.Fatal("NewCoverageReport returned nil")
	}

	if report.Files == nil {
		t.Error("Files map is nil")
	}

	if len(report.Files) != 0 {
		t.Errorf("Expected empty Files map, got %d files", len(report.Files))
	}
}

func TestAddFile(t *testing.T) {
	report := NewCoverageReport()

	fc := &FileCoverage{
		FileName:     "test.go",
		TotalLines:   100,
		CoveredLines: 80,
		CoveragePct:  80.0,
	}

	report.AddFile(fc)

	if len(report.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(report.Files))
	}

	retrieved := report.Files["test.go"]
	if retrieved == nil {
		t.Fatal("File not found in report")
	}

	if retrieved.FileName != "test.go" {
		t.Errorf("Expected filename 'test.go', got '%s'", retrieved.FileName)
	}
}

func TestGetFile(t *testing.T) {
	report := NewCoverageReport()

	fc := &FileCoverage{
		FileName:     "main.go",
		TotalLines:   50,
		CoveredLines: 40,
	}

	report.AddFile(fc)

	// Test getting existing file
	retrieved := report.GetFile("main.go")
	if retrieved == nil {
		t.Fatal("GetFile returned nil for existing file")
	}

	if retrieved.FileName != "main.go" {
		t.Errorf("Expected 'main.go', got '%s'", retrieved.FileName)
	}

	// Test getting non-existing file
	notFound := report.GetFile("nonexistent.go")
	if notFound != nil {
		t.Error("Expected nil for non-existent file")
	}
}

func TestCalculateCoverage(t *testing.T) {
	tests := []struct {
		name         string
		totalLines   int
		coveredLines int
		expectedPct  float64
	}{
		{
			name:         "80% coverage",
			totalLines:   100,
			coveredLines: 80,
			expectedPct:  80.0,
		},
		{
			name:         "50% coverage",
			totalLines:   200,
			coveredLines: 100,
			expectedPct:  50.0,
		},
		{
			name:         "100% coverage",
			totalLines:   50,
			coveredLines: 50,
			expectedPct:  100.0,
		},
		{
			name:         "0% coverage",
			totalLines:   100,
			coveredLines: 0,
			expectedPct:  0.0,
		},
		{
			name:         "Zero lines",
			totalLines:   0,
			coveredLines: 0,
			expectedPct:  0.0,
		},
		{
			name:         "Partial coverage",
			totalLines:   3,
			coveredLines: 2,
			expectedPct:  66.66666666666667,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FileCoverage{
				FileName:     "test.go",
				TotalLines:   tt.totalLines,
				CoveredLines: tt.coveredLines,
			}

			fc.CalculateCoverage()

			// Use approximate comparison for floating point
			diff := fc.CoveragePct - tt.expectedPct
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.0001 {
				t.Errorf("Expected coverage %.2f%%, got %.2f%%", tt.expectedPct, fc.CoveragePct)
			}
		})
	}
}

func TestCoverageReportWithMultipleFiles(t *testing.T) {
	report := NewCoverageReport()
	report.TestName = "Integration Tests"

	files := []*FileCoverage{
		{FileName: "file1.go", TotalLines: 100, CoveredLines: 90},
		{FileName: "file2.go", TotalLines: 50, CoveredLines: 40},
		{FileName: "file3.go", TotalLines: 75, CoveredLines: 60},
	}

	for _, fc := range files {
		fc.CalculateCoverage()
		report.AddFile(fc)
	}

	if len(report.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(report.Files))
	}

	if report.TestName != "Integration Tests" {
		t.Errorf("Expected test name 'Integration Tests', got '%s'", report.TestName)
	}

	// Verify each file
	file1 := report.GetFile("file1.go")
	if file1 == nil || file1.CoveragePct != 90.0 {
		t.Error("file1.go coverage incorrect")
	}

	file2 := report.GetFile("file2.go")
	if file2 == nil || file2.CoveragePct != 80.0 {
		t.Error("file2.go coverage incorrect")
	}

	file3 := report.GetFile("file3.go")
	if file3 == nil || file3.CoveragePct != 80.0 {
		t.Error("file3.go coverage incorrect")
	}
}

func TestCalculateOverallCoverage(t *testing.T) {
	tests := []struct {
		name                 string
		files                []*FileCoverage
		expectedTotalLines   int
		expectedTotalCovered int
		expectedOverallPct   float64
	}{
		{
			name: "Multiple files with mixed coverage",
			files: []*FileCoverage{
				{FileName: "file1.go", TotalLines: 100, CoveredLines: 90},
				{FileName: "file2.go", TotalLines: 50, CoveredLines: 40},
				{FileName: "file3.go", TotalLines: 75, CoveredLines: 60},
			},
			expectedTotalLines:   225,
			expectedTotalCovered: 190,
			expectedOverallPct:   (190.0 / 225.0) * 100.0, // ~84.44%
		},
		{
			name: "Single file",
			files: []*FileCoverage{
				{FileName: "single.go", TotalLines: 200, CoveredLines: 150},
			},
			expectedTotalLines:   200,
			expectedTotalCovered: 150,
			expectedOverallPct:   75.0,
		},
		{
			name:                 "Empty report",
			files:                []*FileCoverage{},
			expectedTotalLines:   0,
			expectedTotalCovered: 0,
			expectedOverallPct:   0.0,
		},
		{
			name: "Files with zero total lines",
			files: []*FileCoverage{
				{FileName: "empty1.go", TotalLines: 0, CoveredLines: 0},
				{FileName: "empty2.go", TotalLines: 0, CoveredLines: 0},
			},
			expectedTotalLines:   0,
			expectedTotalCovered: 0,
			expectedOverallPct:   0.0,
		},
		{
			name: "Mixed files including zero lines",
			files: []*FileCoverage{
				{FileName: "normal.go", TotalLines: 100, CoveredLines: 80},
				{FileName: "empty.go", TotalLines: 0, CoveredLines: 0},
			},
			expectedTotalLines:   100,
			expectedTotalCovered: 80,
			expectedOverallPct:   80.0,
		},
		{
			name: "All lines covered",
			files: []*FileCoverage{
				{FileName: "perfect1.go", TotalLines: 50, CoveredLines: 50},
				{FileName: "perfect2.go", TotalLines: 30, CoveredLines: 30},
			},
			expectedTotalLines:   80,
			expectedTotalCovered: 80,
			expectedOverallPct:   100.0,
		},
		{
			name: "No lines covered",
			files: []*FileCoverage{
				{FileName: "uncovered1.go", TotalLines: 40, CoveredLines: 0},
				{FileName: "uncovered2.go", TotalLines: 60, CoveredLines: 0},
			},
			expectedTotalLines:   100,
			expectedTotalCovered: 0,
			expectedOverallPct:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewCoverageReport()
			for _, fc := range tt.files {
				report.AddFile(fc)
			}

			totalLines, totalCovered, overallPct := report.CalculateOverallCoverage()

			if totalLines != tt.expectedTotalLines {
				t.Errorf("Expected total lines %d, got %d", tt.expectedTotalLines, totalLines)
			}

			if totalCovered != tt.expectedTotalCovered {
				t.Errorf("Expected total covered %d, got %d", tt.expectedTotalCovered, totalCovered)
			}

			// Use approximate comparison for floating point
			diff := overallPct - tt.expectedOverallPct
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.0001 {
				t.Errorf("Expected overall coverage %.4f%%, got %.4f%%", tt.expectedOverallPct, overallPct)
			}
		})
	}
}

func TestFileCoverageWithFunctions(t *testing.T) {
	fc := &FileCoverage{
		FileName:     "utils.go",
		TotalLines:   100,
		CoveredLines: 85,
		Functions: []FunctionCoverage{
			{Name: "Add", LineNumber: 10, ExecutionCount: 5},
			{Name: "Subtract", LineNumber: 20, ExecutionCount: 3},
			{Name: "Multiply", LineNumber: 30, ExecutionCount: 0},
		},
	}

	fc.CalculateCoverage()

	if len(fc.Functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(fc.Functions))
	}

	if fc.Functions[0].Name != "Add" {
		t.Errorf("Expected first function 'Add', got '%s'", fc.Functions[0].Name)
	}

	if fc.Functions[2].ExecutionCount != 0 {
		t.Error("Expected Multiply function to have 0 executions")
	}
}

func TestFileCoverageWithLines(t *testing.T) {
	fc := &FileCoverage{
		FileName:     "main.go",
		TotalLines:   10,
		CoveredLines: 8,
		Lines: map[int]LineCoverage{
			1: {LineNumber: 1, ExecutionCount: 1, Checksum: "abc123"},
			2: {LineNumber: 2, ExecutionCount: 5, Checksum: "def456"},
			3: {LineNumber: 3, ExecutionCount: 0, Checksum: "ghi789"},
		},
	}

	fc.CalculateCoverage()

	if len(fc.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(fc.Lines))
	}

	line1 := fc.Lines[1]
	if line1.ExecutionCount != 1 {
		t.Errorf("Expected line 1 execution count 1, got %d", line1.ExecutionCount)
	}

	line2 := fc.Lines[2]
	if line2.Checksum != "def456" {
		t.Errorf("Expected checksum 'def456', got '%s'", line2.Checksum)
	}

	line3 := fc.Lines[3]
	if line3.ExecutionCount != 0 {
		t.Error("Expected line 3 to be uncovered")
	}
}
