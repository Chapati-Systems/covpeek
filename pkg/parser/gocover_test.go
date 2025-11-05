package parser

import (
	"strings"
	"testing"
)

func TestGoCoverParser_Parse_ValidFile(t *testing.T) {
	input := `mode: set
myproject/file.go:5.10,7.2 2 1
myproject/file.go:9.15,11.2 1 0
`
	
	parser := NewGoCoverParser()
	report, err := parser.Parse(strings.NewReader(input))
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if parser.GetMode() != "set" {
		t.Errorf("Expected mode 'set', got: %s", parser.GetMode())
	}
	
	if len(report.Files) != 1 {
		t.Fatalf("Expected 1 file, got: %d", len(report.Files))
	}
	
	file := report.Files["myproject/file.go"]
	if file == nil {
		t.Fatalf("Expected file 'myproject/file.go' not found")
	}
	
	if file.TotalLines == 0 {
		t.Error("Expected non-zero total lines")
	}
	
	// Check that lines 5-7 are covered (count 1)
	for i := 5; i <= 7; i++ {
		if line, exists := file.Lines[i]; !exists || line.ExecutionCount != 1 {
			t.Errorf("Expected line %d to be covered with count 1", i)
		}
	}
	
	// Check that lines 9-11 are not covered (count 0)
	for i := 9; i <= 11; i++ {
		if line, exists := file.Lines[i]; !exists || line.ExecutionCount != 0 {
			t.Errorf("Expected line %d to have count 0", i)
		}
	}
}

func TestGoCoverParser_Parse_CountMode(t *testing.T) {
	input := `mode: count
myproject/file.go:5.10,7.2 1 3
myproject/file.go:5.10,7.2 1 2
`
	
	parser := NewGoCoverParser()
	report, err := parser.Parse(strings.NewReader(input))
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if parser.GetMode() != "count" {
		t.Errorf("Expected mode 'count', got: %s", parser.GetMode())
	}
	
	file := report.Files["myproject/file.go"]
	if file == nil {
		t.Fatalf("Expected file 'myproject/file.go' not found")
	}
	
	// In count mode, overlapping entries should sum
	for i := 5; i <= 7; i++ {
		line := file.Lines[i]
		if line.ExecutionCount != 5 { // 3 + 2
			t.Errorf("Expected line %d execution count 5, got: %d", i, line.ExecutionCount)
		}
	}
}

func TestGoCoverParser_Parse_MultipleFiles(t *testing.T) {
	input := `mode: set
file1.go:1.1,3.2 1 1
file2.go:1.1,2.2 1 0
`
	
	parser := NewGoCoverParser()
	report, err := parser.Parse(strings.NewReader(input))
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(report.Files) != 2 {
		t.Fatalf("Expected 2 files, got: %d", len(report.Files))
	}
	
	file1 := report.Files["file1.go"]
	if file1 == nil {
		t.Fatalf("Expected file 'file1.go' not found")
	}
	
	file2 := report.Files["file2.go"]
	if file2 == nil {
		t.Fatalf("Expected file 'file2.go' not found")
	}
}

func TestGoCoverParser_Parse_EmptyFile(t *testing.T) {
	input := `mode: set
`
	
	parser := NewGoCoverParser()
	report, err := parser.Parse(strings.NewReader(input))
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(report.Files) != 0 {
		t.Errorf("Expected 0 files, got: %d", len(report.Files))
	}
}

func TestGoCoverParser_Parse_MissingMode(t *testing.T) {
	input := `file.go:1.1,2.2 1 1
`
	
	parser := NewGoCoverParser()
	_, err := parser.Parse(strings.NewReader(input))
	
	if err == nil {
		t.Error("Expected error for missing mode declaration, got nil")
	}
}

func TestGoCoverParser_Parse_MalformedEntry(t *testing.T) {
	input := `mode: set
file.go:invalid_entry
file.go:5.10,7.2 1 1
`
	
	parser := NewGoCoverParser()
	report, err := parser.Parse(strings.NewReader(input))
	
	// Should not error but should generate warnings
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warnings for malformed entry, got none")
	}
	
	// Should still parse valid entries
	file := report.Files["file.go"]
	if file == nil {
		t.Fatalf("Expected file 'file.go' not found")
	}
}

func TestGoCoverParser_Parse_AtomicMode(t *testing.T) {
	input := `mode: atomic
file.go:1.1,3.2 1 5
`
	
	parser := NewGoCoverParser()
	report, err := parser.Parse(strings.NewReader(input))
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if parser.GetMode() != "atomic" {
		t.Errorf("Expected mode 'atomic', got: %s", parser.GetMode())
	}
	
	file := report.Files["file.go"]
	if file == nil {
		t.Fatalf("Expected file 'file.go' not found")
	}
}

func TestGoCoverParser_Parse_CoverageCalculation(t *testing.T) {
	input := `mode: set
file.go:1.1,3.2 1 1
file.go:5.1,7.2 1 0
`
	
	parser := NewGoCoverParser()
	report, err := parser.Parse(strings.NewReader(input))
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	file := report.Files["file.go"]
	if file == nil {
		t.Fatalf("Expected file 'file.go' not found")
	}
	
	// Coverage should be calculated
	if file.CoveragePct == 0 {
		t.Error("Expected non-zero coverage percentage")
	}
	
	// Should have 6 lines total (1-3, 5-7)
	if file.TotalLines != 6 {
		t.Errorf("Expected 6 total lines, got: %d", file.TotalLines)
	}
	
	// Should have 3 covered lines (1-3)
	if file.CoveredLines != 3 {
		t.Errorf("Expected 3 covered lines, got: %d", file.CoveredLines)
	}
	
	expectedCoverage := 50.0
	if file.CoveragePct < expectedCoverage-0.1 || file.CoveragePct > expectedCoverage+0.1 {
		t.Errorf("Expected coverage ~%.2f%%, got: %.2f%%", expectedCoverage, file.CoveragePct)
	}
}
