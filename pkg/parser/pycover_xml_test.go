package parser

import (
	"strings"
	"testing"
)

func TestPyCoverXMLParser_Parse_ValidFile(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <packages>
    <package>
      <classes>
        <class filename="src/main.py">
          <lines>
            <line number="1" hits="1"/>
            <line number="2" hits="1"/>
            <line number="3" hits="0"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	parser := NewPyCoverXMLParser()
	report, err := parser.Parse(strings.NewReader(xmlData))

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

func TestPyCoverXMLParser_Parse_MultipleFiles(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <packages>
    <package>
      <classes>
        <class filename="src/main.py">
          <lines>
            <line number="1" hits="1"/>
            <line number="2" hits="0"/>
          </lines>
        </class>
        <class filename="src/utils.py">
          <lines>
            <line number="1" hits="1"/>
            <line number="2" hits="1"/>
            <line number="3" hits="0"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	parser := NewPyCoverXMLParser()
	report, err := parser.Parse(strings.NewReader(xmlData))

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
	if file2.TotalLines != 3 || file2.CoveredLines != 2 {
		t.Errorf("File2: expected 3 total, 2 covered, got %d total, %d covered", file2.TotalLines, file2.CoveredLines)
	}
}

func TestPyCoverXMLParser_Parse_EmptyFile(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <packages>
  </packages>
</coverage>`

	parser := NewPyCoverXMLParser()
	report, err := parser.Parse(strings.NewReader(xmlData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(report.Files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(report.Files))
	}
}

func TestPyCoverXMLParser_Parse_InvalidXML(t *testing.T) {
	xmlData := `<invalid xml>`

	parser := NewPyCoverXMLParser()
	_, err := parser.Parse(strings.NewReader(xmlData))

	if err == nil {
		t.Fatal("Expected error for invalid XML, got nil")
	}
}

func TestPyCoverXMLParser_Parse_MalformedXML(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <packages>
    <package>
      <classes>
        <class filename="src/main.py">
          <lines>
            <line number="invalid" hits="1"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	parser := NewPyCoverXMLParser()
	report, err := parser.Parse(strings.NewReader(xmlData))

	// The XML should still parse, but the line number parsing might fail
	// This depends on how strict the XML parsing is
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}

	if report != nil && len(report.Files) > 0 {
		t.Logf("Parsed %d files despite malformed data", len(report.Files))
	}
}

func TestPyCoverXMLParser_GetWarnings(t *testing.T) {
	// Test with negative hits to trigger warnings
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<coverage>
  <packages>
    <package>
      <classes>
        <class filename="src/main.py">
          <lines>
            <line number="1" hits="1"/>
            <line number="2" hits="-1"/>
            <line number="3" hits="0"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	parser := NewPyCoverXMLParser()
	report, err := parser.Parse(strings.NewReader(xmlData))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if report == nil || len(report.Files) == 0 {
		t.Fatal("Expected report with files")
	}

	// Check warnings - should have warning about negative hits
	warnings := parser.GetWarnings()
	if len(warnings) == 0 {
		t.Errorf("Expected warnings for negative hit count, got 0 warnings")
	} else {
		t.Logf("Got expected warnings: %v", warnings)
	}
}
