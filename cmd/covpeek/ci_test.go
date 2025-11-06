package main

import (
	"testing"

	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
)

func TestMergeReports(t *testing.T) {
	// Create two reports
	report1 := models.NewCoverageReport()
	file1 := &models.FileCoverage{
		FileName:     "file1.go",
		TotalLines:   100,
		CoveredLines: 80,
	}
	file1.CalculateCoverage()
	report1.AddFile(file1)

	report2 := models.NewCoverageReport()
	file2 := &models.FileCoverage{
		FileName:     "file2.go",
		TotalLines:   50,
		CoveredLines: 40,
	}
	file2.CalculateCoverage()
	report2.AddFile(file2)

	reports := []*models.CoverageReport{report1, report2}
	merged := mergeReports(reports)

	// Check merged
	if len(merged.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(merged.Files))
	}

	_, _, pct := merged.CalculateOverallCoverage()
	expectedPct := (80.0 + 40.0) / (100.0 + 50.0) * 100.0
	if pct != expectedPct {
		t.Errorf("Expected %.2f%%, got %.2f%%", expectedPct, pct)
	}
}

func TestMergeReportsOverlappingFiles(t *testing.T) {
	// Create reports with same file
	report1 := models.NewCoverageReport()
	file1 := &models.FileCoverage{
		FileName:     "file.go",
		TotalLines:   100,
		CoveredLines: 80,
	}
	file1.CalculateCoverage()
	report1.AddFile(file1)

	report2 := models.NewCoverageReport()
	file2 := &models.FileCoverage{
		FileName:     "file.go",
		TotalLines:   50,
		CoveredLines: 40,
	}
	file2.CalculateCoverage()
	report2.AddFile(file2)

	reports := []*models.CoverageReport{report1, report2}
	merged := mergeReports(reports)

	// Check merged
	if len(merged.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(merged.Files))
	}

	fc := merged.GetFile("file.go")
	if fc.TotalLines != 150 || fc.CoveredLines != 120 {
		t.Errorf("Expected TotalLines 150, CoveredLines 120, got %d, %d", fc.TotalLines, fc.CoveredLines)
	}

	expectedPct := 120.0 / 150.0 * 100.0
	if fc.CoveragePct != expectedPct {
		t.Errorf("Expected %.2f%%, got %.2f%%", expectedPct, fc.CoveragePct)
	}
}

func TestParseCoverageFile(t *testing.T) {
	// Test with existing testdata
	report, err := parseCoverageFile("../../testdata/sample.lcov")
	if err != nil {
		t.Errorf("Failed to parse sample.lcov: %v", err)
	}

	if report == nil {
		t.Error("Report is nil")
	}

	// Check has files
	if len(report.Files) == 0 {
		t.Error("No files in report")
	}
}

func TestParseCoverageFileNonExistent(t *testing.T) {
	_, err := parseCoverageFile("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestMergeReportsEmpty(t *testing.T) {
	reports := []*models.CoverageReport{}
	merged := mergeReports(reports)

	if merged == nil {
		t.Error("Expected non-nil merged report")
	}

	if len(merged.Files) != 0 {
		t.Error("Expected 0 files")
	}
}

func TestMergeReportsSingle(t *testing.T) {
	report := models.NewCoverageReport()
	file := &models.FileCoverage{
		FileName:     "file.go",
		TotalLines:   100,
		CoveredLines: 80,
	}
	file.CalculateCoverage()
	report.AddFile(file)

	reports := []*models.CoverageReport{report}
	merged := mergeReports(reports)

	if len(merged.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(merged.Files))
	}

	fc := merged.GetFile("file.go")
	if fc.TotalLines != 100 || fc.CoveredLines != 80 {
		t.Errorf("Expected TotalLines 100, CoveredLines 80, got %d, %d", fc.TotalLines, fc.CoveredLines)
	}
}
