package parser

import (
	"strings"
	"testing"
)

func TestLCOVParser_Parse_ValidFile(t *testing.T) {
	input := `TN:test_name
SF:src/lib.rs
FN:5,my_function
FNDA:3,my_function
FNF:1
FNH:1
DA:5,3
DA:6,3
DA:7,0
LH:2
LF:3
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report.TestName != "test_name" {
		t.Errorf("Expected test name 'test_name', got: %s", report.TestName)
	}

	if len(report.Files) != 1 {
		t.Fatalf("Expected 1 file, got: %d", len(report.Files))
	}

	file := report.Files["src/lib.rs"]
	if file == nil {
		t.Fatalf("Expected file 'src/lib.rs' not found")
	}

	if file.TotalLines != 3 {
		t.Errorf("Expected 3 total lines, got: %d", file.TotalLines)
	}

	if file.CoveredLines != 2 {
		t.Errorf("Expected 2 covered lines, got: %d", file.CoveredLines)
	}

	expectedCoverage := 66.67
	if file.CoveragePct < expectedCoverage-0.1 || file.CoveragePct > expectedCoverage+0.1 {
		t.Errorf("Expected coverage ~%.2f%%, got: %.2f%%", expectedCoverage, file.CoveragePct)
	}

	if len(file.Functions) != 1 {
		t.Fatalf("Expected 1 function, got: %d", len(file.Functions))
	}

	if file.Functions[0].Name != "my_function" {
		t.Errorf("Expected function name 'my_function', got: %s", file.Functions[0].Name)
	}

	if file.Functions[0].ExecutionCount != 3 {
		t.Errorf("Expected function execution count 3, got: %d", file.Functions[0].ExecutionCount)
	}
}

func TestLCOVParser_Parse_MultipleFiles(t *testing.T) {
	input := `TN:multi_test
SF:file1.rs
DA:1,1
DA:2,1
LH:2
LF:2
end_of_record
SF:file2.rs
DA:1,0
DA:2,0
LH:0
LF:2
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(report.Files) != 2 {
		t.Fatalf("Expected 2 files, got: %d", len(report.Files))
	}

	file1 := report.Files["file1.rs"]
	if file1 == nil {
		t.Fatalf("Expected file 'file1.rs' not found")
	}

	if file1.CoveragePct != 100.0 {
		t.Errorf("Expected file1 coverage 100%%, got: %.2f%%", file1.CoveragePct)
	}

	file2 := report.Files["file2.rs"]
	if file2 == nil {
		t.Fatalf("Expected file 'file2.rs' not found")
	}

	if file2.CoveragePct != 0.0 {
		t.Errorf("Expected file2 coverage 0%%, got: %.2f%%", file2.CoveragePct)
	}
}

func TestLCOVParser_Parse_MalformedLines(t *testing.T) {
	input := `TN:test
SF:file.rs
DA:invalid_line
DA:5,10
LH:1
LF:1
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	// Should not error, but should log warnings
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that we collected warnings
	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warnings for malformed line, got none")
	}

	// Should still have parsed valid lines
	if len(report.Files) != 1 {
		t.Fatalf("Expected 1 file, got: %d", len(report.Files))
	}
}

func TestLCOVParser_Parse_EmptyFile(t *testing.T) {
	input := ``

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(report.Files) != 0 {
		t.Errorf("Expected 0 files, got: %d", len(report.Files))
	}
}

func TestLCOVParser_Parse_WithChecksum(t *testing.T) {
	input := `SF:file.rs
DA:1,5,abc123def456
DA:2,3
LH:2
LF:2
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	file := report.Files["file.rs"]
	if file == nil {
		t.Fatalf("Expected file 'file.rs' not found")
	}

	line1 := file.Lines[1]
	if line1.Checksum != "abc123def456" {
		t.Errorf("Expected checksum 'abc123def456', got: %s", line1.Checksum)
	}

	line2 := file.Lines[2]
	if line2.Checksum != "" {
		t.Errorf("Expected empty checksum, got: %s", line2.Checksum)
	}
}

func TestLCOVParser_Parse_RecordWithoutSourceFile(t *testing.T) {
	input := `TN:test
DA:1,1
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	// Should not error, but should generate warnings
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warnings for DA without SF, got none")
	}

	if len(report.Files) != 0 {
		t.Errorf("Expected 0 files, got: %d", len(report.Files))
	}
}

func TestLCOVParser_Parse_BranchCoverage(t *testing.T) {
	input := `SF:file.rs
BRF:10
BRH:7
BRDA:1,0,0,5
BRDA:1,0,1,3
DA:1,8
LH:1
LF:1
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Branch coverage records should be gracefully skipped
	file := report.Files["file.rs"]
	if file == nil {
		t.Fatalf("Expected file 'file.rs' not found")
	}
}

func TestLCOVParser_Parse_UnknownRecordType(t *testing.T) {
	input := `SF:file.rs
UNKNOWN:some_value
DA:1,1
LH:1
LF:1
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	// Should not error, but should log warning
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warning for unknown record type, got none")
	}

	// Should still parse valid records
	file := report.Files["file.rs"]
	if file == nil {
		t.Fatalf("Expected file 'file.rs' not found")
	}
}

func TestLCOVParser_Parse_FNF_FNH(t *testing.T) {
	input := `SF:file.rs
FN:1,func1
FN:5,func2
FNDA:1,func1
FNDA:0,func2
FNF:2
FNH:1
DA:1,1
LH:1
LF:1
end_of_record
`

	parser := NewLCOVParser()
	report, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// FNF and FNH are parsed but not validated
	file := report.Files["file.rs"]
	if file == nil {
		t.Fatalf("Expected file 'file.rs' not found")
	}

	if len(file.Functions) != 2 {
		t.Errorf("Expected 2 functions, got: %d", len(file.Functions))
	}
}

func TestLCOVParser_Parse_InvalidFN(t *testing.T) {
	input := `SF:file.rs
FN:invalid
DA:1,1
LH:1
LF:1
end_of_record
`

	parser := NewLCOVParser()
	_, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warning for invalid FN, got none")
	}
}

func TestLCOVParser_Parse_InvalidFNDA(t *testing.T) {
	input := `SF:file.rs
FNDA:invalid,func
DA:1,1
LH:1
LF:1
end_of_record
`

	parser := NewLCOVParser()
	_, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warning for invalid FNDA, got none")
	}
}

func TestLCOVParser_Parse_InvalidLH(t *testing.T) {
	input := `SF:file.rs
DA:1,1
LH:invalid
LF:1
end_of_record
`

	parser := NewLCOVParser()
	_, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warning for invalid LH, got none")
	}
}

func TestLCOVParser_Parse_InvalidLF(t *testing.T) {
	input := `SF:file.rs
DA:1,1
LH:1
LF:invalid
end_of_record
`

	parser := NewLCOVParser()
	_, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warning for invalid LF, got none")
	}
}
