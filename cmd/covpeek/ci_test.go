package main

import (
	"os"
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

func TestGetPossibleCoverageFiles(t *testing.T) {
	files := getPossibleCoverageFiles()
	expected := []string{
		"coverage.out",
		"test/coverage.out",
		"lcov.info",
		"target/coverage/lcov.info",
		"coverage/lcov.info",
		"coverage/coverage-final.json",
		"coverage.xml",
		"coverage.json",
	}

	if len(files) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(files))
	}

	for i, file := range files {
		if file != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], file)
		}
	}
}

func TestDetectExistingCoverageFiles(t *testing.T) {
	// Create a temporary file
	tmpFile := "coverage.out"
	err := os.WriteFile(tmpFile, []byte("mode: set\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	files := detectExistingCoverageFiles()
	if len(files) == 0 {
		t.Error("Expected at least one existing file")
	}

	found := false
	for _, file := range files {
		if file == tmpFile {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find %s in existing files", tmpFile)
	}
}

func TestRunCI(t *testing.T) {
	// Create a temporary coverage file
	tmpFile := "coverage.out"
	content := `mode: set
github.com/example/main.go:10.1,12.1 1 1
github.com/example/main.go:15.1,17.1 1 0
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Test with min coverage that should pass
	minCoverage = 50.0

	// We can't easily test the exit codes without refactoring, so just test the logic
	// by calling the functions directly
	existingFiles := detectExistingCoverageFiles()
	if len(existingFiles) == 0 {
		t.Error("Expected to detect coverage file")
	}

	report, err := parseCoverageFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to parse coverage file: %v", err)
	}

	if report == nil {
		t.Error("Report is nil")
	}

	_, _, pct := report.CalculateOverallCoverage()
	if pct <= 0 {
		t.Error("Expected positive coverage percentage")
	}
}

func TestParseCoverageFileGo(t *testing.T) {
	report, err := parseCoverageFile("../../testdata/sample.out")
	if err != nil {
		t.Errorf("Failed to parse sample.out: %v", err)
	}

	if report == nil {
		t.Error("Report is nil")
	}

	if len(report.Files) == 0 {
		t.Error("No files in report")
	}
}

func TestParseCoverageFilePyJSON(t *testing.T) {
	report, err := parseCoverageFile("../../testdata/coverage.json")
	if err != nil {
		t.Errorf("Failed to parse coverage.json: %v", err)
	}

	if report == nil {
		t.Error("Report is nil")
	}

	if len(report.Files) == 0 {
		t.Error("No files in report")
	}
}

func TestParseCoverageFilePyXML(t *testing.T) {
	report, err := parseCoverageFile("../../testdata/coverage.xml")
	if err != nil {
		t.Errorf("Failed to parse coverage.xml: %v", err)
	}

	if report == nil {
		t.Error("Report is nil")
	}

	if len(report.Files) == 0 {
		t.Error("No files in report")
	}
}
