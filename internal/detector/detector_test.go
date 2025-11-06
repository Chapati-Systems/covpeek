package detector

import (
	"strings"
	"testing"
)

func TestDetectFormat_LCOV(t *testing.T) {
	input := `TN:test
SF:file.rs
DA:1,1
end_of_record
`

	format, err := DetectFormat(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if format != LCOVFormat {
		t.Errorf("Expected LCOVFormat, got: %s", format)
	}
}

func TestDetectFormat_GoCoverage(t *testing.T) {
	input := `mode: set
file.go:1.1,3.2 1 1
`

	format, err := DetectFormat(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if format != GoCoverFormat {
		t.Errorf("Expected GoCoverFormat, got: %s", format)
	}
}

func TestDetectFormat_Unknown(t *testing.T) {
	input := `some random content
that doesn't match any format
`

	format, err := DetectFormat(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if format != UnknownFormat {
		t.Errorf("Expected UnknownFormat, got: %s", format)
	}
}

func TestDetectFormat_Empty(t *testing.T) {
	input := ``

	format, err := DetectFormat(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if format != UnknownFormat {
		t.Errorf("Expected UnknownFormat, got: %s", format)
	}
}

func TestDetectFormatByExtension_LCOV(t *testing.T) {
	tests := []string{
		"coverage.lcov",
		"test.info",
		"lcov.info",
		"coverage/lcov.info",
		"COVERAGE.LCOV", // Test case insensitive
	}

	for _, filename := range tests {
		format := DetectFormatByExtension(filename)
		if format != LCOVFormat {
			t.Errorf("Expected LCOVFormat for %s, got: %s", filename, format)
		}
	}
}

func TestDetectFormatByExtension_Go(t *testing.T) {
	tests := []string{
		"coverage.out",
		"test/coverage.out",
		"COVERAGE.OUT", // Test case insensitive
	}

	for _, filename := range tests {
		format := DetectFormatByExtension(filename)
		if format != GoCoverFormat {
			t.Errorf("Expected GoCoverFormat for %s, got: %s", filename, format)
		}
	}
}

func TestDetectFormatByExtension_PyXML(t *testing.T) {
	tests := []string{
		"coverage.xml",
		"test/coverage.xml",
		"COVERAGE.XML", // Test case insensitive
	}

	for _, filename := range tests {
		format := DetectFormatByExtension(filename)
		if format != PyCoverXMLFormat {
			t.Errorf("Expected PyCoverXMLFormat for %s, got: %s", filename, format)
		}
	}
}

func TestDetectFormatByExtension_PyJSON(t *testing.T) {
	tests := []string{
		"coverage.json",
		"test/coverage.json",
		"COVERAGE.JSON", // Test case insensitive
	}

	for _, filename := range tests {
		format := DetectFormatByExtension(filename)
		if format != PyCoverJSONFormat {
			t.Errorf("Expected PyCoverJSONFormat for %s, got: %s", filename, format)
		}
	}
}

func TestCoverageFormat_String(t *testing.T) {
	tests := []struct {
		format   CoverageFormat
		expected string
	}{
		{LCOVFormat, "LCOV"},
		{GoCoverFormat, "Go Coverage"},
		{PyCoverXMLFormat, "Python XML Coverage"},
		{PyCoverJSONFormat, "Python JSON Coverage"},
		{UnknownFormat, "Unknown"},
	}

	for _, test := range tests {
		result := test.format.String()
		if result != test.expected {
			t.Errorf("Expected %s, got: %s", test.expected, result)
		}
	}
}

func TestDetectFormat_MixedMarkers(t *testing.T) {
	// Should detect Go format even with some non-Go content after
	input := `mode: set
file.go:1.1,3.2 1 1
some other content
`

	format, err := DetectFormat(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if format != GoCoverFormat {
		t.Errorf("Expected GoCoverFormat, got: %s", format)
	}
}

func TestDetectFormat_PyCoverJSON(t *testing.T) {
	input := `{
  "meta": {
    "version": "5.0"
  },
  "files": {
    "file.py": {
      "summary": {
        "covered_lines": 10,
        "num_statements": 12
      }
    }
  }
}`

	format, err := DetectFormat(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if format != PyCoverJSONFormat {
		t.Errorf("Expected PyCoverJSONFormat, got: %s", format)
	}
}

func TestDetectFormat_PyCoverXML(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<coverage version="5.0">
  <sources>
    <source>.</source>
  </sources>
  <packages>
    <package name="">
      <classes>
        <class filename="file.py" name="file" complexity="0">
          <methods/>
          <lines>
            <line hits="1" number="1"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	format, err := DetectFormat(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if format != PyCoverXMLFormat {
		t.Errorf("Expected PyCoverXMLFormat, got: %s", format)
	}
}
