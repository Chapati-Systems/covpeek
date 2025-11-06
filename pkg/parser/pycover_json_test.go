package parser

import (
	"strings"
	"testing"
)

func TestPyCoverJSONParser_Parse_ValidFile(t *testing.T) {
	jsonData := `{
  "files": {
    "src/main.py": {
      "executed_lines": [1, 2],
      "missing_lines": [3],
      "summary": {
        "covered_lines": 2,
        "num_statements": 3,
        "percent_covered": 66.67
      }
    }
  }
}`

	parser := NewPyCoverJSONParser()
	report, err := parser.Parse(strings.NewReader(jsonData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report == nil {
		t.Fatal("Expected report to be non-nil")
	}

	if len(report.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(report.Files))
	}

	file, exists := report.Files["src/main.py"]
	if !exists {
		t.Fatal("Expected file 'src/main.py' to exist")
	}

	if file.TotalLines != 3 {
		t.Errorf("Expected 3 total lines, got %d", file.TotalLines)
	}

	if file.CoveredLines != 2 {
		t.Errorf("Expected 2 covered lines, got %d", file.CoveredLines)
	}

	if file.CoveragePct != 66.66666666666666 {
		t.Errorf("Expected coverage percentage 66.67, got %f", file.CoveragePct)
	}
}

func TestPyCoverJSONParser_Parse_MultipleFiles(t *testing.T) {
	jsonData := `{
  "files": {
    "src/main.py": {
      "executed_lines": [1],
      "missing_lines": [2],
      "summary": {
        "covered_lines": 1,
        "num_statements": 2
      }
    },
    "src/utils.py": {
      "executed_lines": [1, 2, 3],
      "missing_lines": [4, 5],
      "summary": {
        "covered_lines": 3,
        "num_statements": 5
      }
    }
  }
}`

	parser := NewPyCoverJSONParser()
	report, err := parser.Parse(strings.NewReader(jsonData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(report.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(report.Files))
	}

	// Check first file
	file1, exists := report.Files["src/main.py"]
	if !exists {
		t.Fatal("Expected file 'src/main.py' to exist")
	}
	if file1.TotalLines != 2 || file1.CoveredLines != 1 {
		t.Errorf("File1: expected 2 total, 1 covered, got %d total, %d covered", file1.TotalLines, file1.CoveredLines)
	}

	// Check second file
	file2, exists := report.Files["src/utils.py"]
	if !exists {
		t.Fatal("Expected file 'src/utils.py' to exist")
	}
	if file2.TotalLines != 5 || file2.CoveredLines != 3 {
		t.Errorf("File2: expected 5 total, 3 covered, got %d total, %d covered", file2.TotalLines, file2.CoveredLines)
	}
}

func TestPyCoverJSONParser_Parse_WithoutSummary(t *testing.T) {
	jsonData := `{
  "files": {
    "src/main.py": {
      "executed_lines": [1, 2],
      "missing_lines": [3, 4]
    }
  }
}`

	parser := NewPyCoverJSONParser()
	report, err := parser.Parse(strings.NewReader(jsonData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	file, exists := report.Files["src/main.py"]
	if !exists {
		t.Fatal("Expected file 'src/main.py' to exist")
	}

	// Should calculate from line data
	if file.TotalLines != 4 {
		t.Errorf("Expected 4 total lines, got %d", file.TotalLines)
	}

	if file.CoveredLines != 2 {
		t.Errorf("Expected 2 covered lines, got %d", file.CoveredLines)
	}
}

func TestPyCoverJSONParser_Parse_EmptyFile(t *testing.T) {
	jsonData := `{"files": {}}`

	parser := NewPyCoverJSONParser()
	report, err := parser.Parse(strings.NewReader(jsonData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(report.Files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(report.Files))
	}
}

func TestPyCoverJSONParser_Parse_InvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`

	parser := NewPyCoverJSONParser()
	_, err := parser.Parse(strings.NewReader(jsonData))

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestPyCoverJSONParser_GetWarnings(t *testing.T) {
	// Test with mismatched summary data to trigger warnings
	jsonData := `{
  "files": {
    "src/main.py": {
      "executed_lines": [1, 2],
      "missing_lines": [3, 4],
      "summary": {
        "covered_lines": 5,
        "num_statements": 10,
        "percent_covered": 50.0,
        "missing_lines": [3, 4],
        "excluded_lines": []
      }
    }
  },
  "totals": {
    "covered_lines": 2,
    "num_statements": 4,
    "percent_covered": 50.0,
    "missing_lines": 2,
    "excluded_lines": 0
  }
}`

	parser := NewPyCoverJSONParser()
	report, err := parser.Parse(strings.NewReader(jsonData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report == nil || len(report.Files) == 0 {
		t.Fatal("Expected report with files")
	}

	// Check warnings - should have warnings about mismatched data
	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("Expected warnings for mismatched summary data, got 0 warnings")
	} else {
		t.Logf("Got expected warnings: %v", warnings)
	}
}
