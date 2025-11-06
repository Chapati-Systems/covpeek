package main

import (
	"os"
	"strings"
	"testing"

	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
)

func TestComputeDiff(t *testing.T) {
	// Create reportA
	reportA := models.NewCoverageReport()
	fcA1 := &models.FileCoverage{
		FileName:     "file1.go",
		TotalLines:   100,
		CoveredLines: 80,
	}
	fcA1.CalculateCoverage()
	reportA.AddFile(fcA1)

	fcA2 := &models.FileCoverage{
		FileName:     "file2.go",
		TotalLines:   50,
		CoveredLines: 40,
	}
	fcA2.CalculateCoverage()
	reportA.AddFile(fcA2)

	// reportB
	reportB := models.NewCoverageReport()
	fcB1 := &models.FileCoverage{
		FileName:     "file1.go",
		TotalLines:   100,
		CoveredLines: 85,
	}
	fcB1.CalculateCoverage()
	reportB.AddFile(fcB1)

	fcB2 := &models.FileCoverage{
		FileName:     "file2.go",
		TotalLines:   50,
		CoveredLines: 35,
	}
	fcB2.CalculateCoverage()
	reportB.AddFile(fcB2)

	fcB3 := &models.FileCoverage{
		FileName:     "file3.go",
		TotalLines:   20,
		CoveredLines: 20,
	}
	fcB3.CalculateCoverage()
	reportB.AddFile(fcB3)

	diff := computeDiff(reportA, reportB)

	// Overall: A: (80+40)/(100+50)=120/150=80%, B: (85+35+20)/(100+50+20)=140/170≈82.35%, delta≈2.35%
	expectedOverallA := 80.0
	expectedOverallB := 140.0 / 170.0 * 100
	expectedDelta := expectedOverallB - expectedOverallA

	if diff.OverallA != expectedOverallA {
		t.Errorf("OverallA: got %.2f, want %.2f", diff.OverallA, expectedOverallA)
	}
	if diff.OverallB < expectedOverallB-0.01 || diff.OverallB > expectedOverallB+0.01 {
		t.Errorf("OverallB: got %.2f, want %.2f", diff.OverallB, expectedOverallB)
	}
	if diff.OverallDelta < expectedDelta-0.01 || diff.OverallDelta > expectedDelta+0.01 {
		t.Errorf("OverallDelta: got %.2f, want %.2f", diff.OverallDelta, expectedDelta)
	}

	// File changes
	expectedFiles := map[string]FileChange{
		"file1.go": {FileName: "file1.go", CoverageA: 80.0, CoverageB: 85.0, Delta: 5.0},
		"file2.go": {FileName: "file2.go", CoverageA: 80.0, CoverageB: 70.0, Delta: -10.0},
		"file3.go": {FileName: "file3.go", CoverageA: 0.0, CoverageB: 100.0, Delta: 100.0},
	}

	if len(diff.FileChanges) != 3 {
		t.Errorf("Expected 3 file changes, got %d", len(diff.FileChanges))
	}

	for _, fc := range diff.FileChanges {
		expected, ok := expectedFiles[fc.FileName]
		if !ok {
			t.Errorf("Unexpected file: %s", fc.FileName)
			continue
		}
		if fc.CoverageA != expected.CoverageA || fc.CoverageB != expected.CoverageB || fc.Delta != expected.Delta {
			t.Errorf("File %s: got %+v, want %+v", fc.FileName, fc, expected)
		}
	}
}

func TestComputeDiffEmptyReports(t *testing.T) {
	reportA := models.NewCoverageReport()
	reportB := models.NewCoverageReport()

	diff := computeDiff(reportA, reportB)

	if diff.OverallA != 0 || diff.OverallB != 0 || diff.OverallDelta != 0 {
		t.Errorf("Expected all zero, got A=%.2f B=%.2f D=%.2f", diff.OverallA, diff.OverallB, diff.OverallDelta)
	}
	if len(diff.FileChanges) != 0 {
		t.Errorf("Expected no file changes, got %d", len(diff.FileChanges))
	}
}

func TestComputeDiffSameReports(t *testing.T) {
	reportA := models.NewCoverageReport()
	fc := &models.FileCoverage{
		FileName:     "file.go",
		TotalLines:   10,
		CoveredLines: 5,
	}
	fc.CalculateCoverage()
	reportA.AddFile(fc)

	reportB := models.NewCoverageReport()
	fc2 := &models.FileCoverage{
		FileName:     "file.go",
		TotalLines:   10,
		CoveredLines: 5,
	}
	fc2.CalculateCoverage()
	reportB.AddFile(fc2)

	diff := computeDiff(reportA, reportB)

	if diff.OverallDelta != 0 {
		t.Errorf("Expected delta 0, got %.2f", diff.OverallDelta)
	}
	if len(diff.FileChanges) != 1 {
		t.Errorf("Expected 1 file change, got %d", len(diff.FileChanges))
	}
	if diff.FileChanges[0].Delta != 0 {
		t.Errorf("Expected file delta 0, got %.2f", diff.FileChanges[0].Delta)
	}
}

func TestComputeDiffOneEmpty(t *testing.T) {
	reportA := models.NewCoverageReport()
	reportB := models.NewCoverageReport()
	fc := &models.FileCoverage{
		FileName:     "file.go",
		TotalLines:   10,
		CoveredLines: 5,
	}
	fc.CalculateCoverage()
	reportB.AddFile(fc)

	diff := computeDiff(reportA, reportB)

	if diff.OverallA != 0 || diff.OverallB != 50 {
		t.Errorf("Expected A=0 B=50, got %.2f %.2f", diff.OverallA, diff.OverallB)
	}
	if len(diff.FileChanges) != 1 {
		t.Errorf("Expected 1 file change, got %d", len(diff.FileChanges))
	}
	if diff.FileChanges[0].CoverageA != 0 || diff.FileChanges[0].CoverageB != 50 {
		t.Errorf("File: expected A=0 B=50, got %.2f %.2f", diff.FileChanges[0].CoverageA, diff.FileChanges[0].CoverageB)
	}
}

func TestRunDiffInvalidOutputFormat(t *testing.T) {
	// Reset flags
	diffFile = ""
	commitA = ""
	commitB = ""
	diffOutputFormat = ""

	rootCmd.SetArgs([]string{"diff", "--file", "test.lcov", "--output", "invalid"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid output format")
	}
}

func TestRunDiffMissingFile(t *testing.T) {
	diffFile = ""
	commitA = ""
	commitB = ""
	diffOutputFormat = ""

	rootCmd.SetArgs([]string{"diff"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for missing --file flag")
	}
}

func TestRunDiffIntegrationMissingFile(t *testing.T) {
	// Assume no coverage.out at HEAD
	diffFile = "nonexistent.lcov"
	commitA = "HEAD~1"
	commitB = "HEAD"
	diffOutputFormat = "detailed"

	rootCmd.SetArgs([]string{"diff", "--file", "nonexistent.lcov"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for missing coverage file at commit")
	}
}

func TestParseCoverageContentLCOV(t *testing.T) {
	content := `TN:test
SF:file.go
DA:1,1
DA:2,0
LH:1
LF:2
end_of_record
`
	report, err := parseCoverageContent([]byte(content), "test.lcov")
	if err != nil {
		t.Fatalf("Failed to parse LCOV: %v", err)
	}
	if report == nil {
		t.Fatal("Report is nil")
	}
	if len(report.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(report.Files))
	}
	fc, ok := report.Files["file.go"]
	if !ok {
		t.Fatal("file.go not found")
	}
	if fc.TotalLines != 2 || fc.CoveredLines != 1 {
		t.Errorf("Expected 2 total, 1 covered, got %d, %d", fc.TotalLines, fc.CoveredLines)
	}
}

func TestParseCoverageContentGo(t *testing.T) {
	content := `mode: set
file.go:1.10,2.20 1 1
file.go:3.30,4.40 0 1
`
	report, err := parseCoverageContent([]byte(content), "coverage.out")
	if err != nil {
		t.Fatalf("Failed to parse Go: %v", err)
	}
	if report == nil {
		t.Fatal("Report is nil")
	}
	// Just check that parsing succeeded
}

func TestParseCoverageContentUnknownFormat(t *testing.T) {
	content := `invalid content`
	_, err := parseCoverageContent([]byte(content), "unknown.txt")
	if err == nil {
		t.Error("Expected error for unknown format")
	}
}

func TestParseCoverageContentPyJSON(t *testing.T) {
	content, err := os.ReadFile("../../testdata/coverage.json")
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}
	report, err := parseCoverageContent(content, "coverage.json")
	if err != nil {
		t.Fatalf("Failed to parse coverage.json: %v", err)
	}
	if report == nil {
		t.Fatal("Report is nil")
	}
	if len(report.Files) == 0 {
		t.Error("Expected files in report")
	}
}

func TestParseCoverageContentPyXML(t *testing.T) {
	content, err := os.ReadFile("../../testdata/coverage.xml")
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}
	report, err := parseCoverageContent(content, "coverage.xml")
	if err != nil {
		t.Fatalf("Failed to parse coverage.xml: %v", err)
	}
	if report == nil {
		t.Fatal("Report is nil")
	}
	if len(report.Files) == 0 {
		t.Error("Expected files in report")
	}
}

func TestOutputDiffSummary(t *testing.T) {
	diff := &CoverageDiff{
		OverallA:     80.0,
		OverallB:     85.0,
		OverallDelta: 5.0,
		FileChanges:  []FileChange{},
	}
	// Just call to cover the code
	outputDiff(diff, "summary")
}

func TestOutputDiffDetailed(t *testing.T) {
	diff := &CoverageDiff{
		OverallA:     80.0,
		OverallB:     85.0,
		OverallDelta: 5.0,
		FileChanges: []FileChange{
			{FileName: "file.go", CoverageA: 70.0, CoverageB: 75.0, Delta: 5.0},
			{FileName: "file2.go", CoverageA: 90.0, CoverageB: 90.0, Delta: 0.0},
		},
	}
	outputDiff(diff, "detailed")
}

func TestOutputDiffJSON(t *testing.T) {
	diff := &CoverageDiff{
		OverallA:     80.0,
		OverallB:     85.0,
		OverallDelta: 5.0,
		FileChanges:  []FileChange{},
	}
	outputDiff(diff, "json")
}

func TestParseCoverageContentMalformed(t *testing.T) {
	content := `invalid content`
	_, err := parseCoverageContent([]byte(content), "unknown.txt")
	if err == nil {
		t.Error("Expected error for malformed content")
	}
}

func TestRunDiffAutoDetectNoFiles(t *testing.T) {
	// Temporarily move existing coverage files
	filesToMove := []string{"coverage.out"}
	movedFiles := make(map[string]string)

	for _, file := range filesToMove {
		if _, err := os.Stat(file); err == nil {
			newName := file + ".backup"
			err := os.Rename(file, newName)
			if err != nil {
				t.Fatalf("Failed to move %s: %v", file, err)
			}
			movedFiles[file] = newName
		}
	}

	// Restore files after test
	defer func() {
		for orig, backup := range movedFiles {
			_ = os.Rename(backup, orig)
		}
	}()

	// Reset flags
	diffFile = ""
	commitA = "HEAD"
	commitB = "HEAD"
	diffOutputFormat = "summary"

	rootCmd.SetArgs([]string{"diff"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for no coverage files detected")
	}
	if !strings.Contains(err.Error(), "no coverage files detected") {
		t.Errorf("Expected 'no coverage files detected' error, got: %v", err)
	}
}

func TestRunDiffAutoDetectMultipleFiles(t *testing.T) {
	// Create multiple coverage files temporarily
	files := []string{"coverage.out", "lcov.info"}
	createdFiles := []string{}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			err := os.WriteFile(file, []byte("dummy"), 0644)
			if err != nil {
				t.Fatalf("Failed to create %s: %v", file, err)
			}
			createdFiles = append(createdFiles, file)
		}
	}

	// Clean up created files
	defer func() {
		for _, file := range createdFiles {
			_ = os.Remove(file)
		}
	}()

	// Reset flags
	diffFile = ""
	commitA = "HEAD"
	commitB = "HEAD"
	diffOutputFormat = "summary"

	rootCmd.SetArgs([]string{"diff"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for multiple coverage files detected")
	}
	if !strings.Contains(err.Error(), "multiple coverage files detected") {
		t.Errorf("Expected 'multiple coverage files detected' error, got: %v", err)
	}
}
