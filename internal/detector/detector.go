package detector

import (
	"bufio"
	"io"
	"strings"
)

// CoverageFormat represents the detected coverage file format
type CoverageFormat int

const (
	// UnknownFormat indicates the format could not be determined
	UnknownFormat CoverageFormat = iota
	// LCOVFormat indicates LCOV format (Rust, TypeScript, JavaScript)
	LCOVFormat
	// GoCoverFormat indicates Go coverage format (.out)
	GoCoverFormat
	// PyCoverXMLFormat indicates Python coverage XML format (Cobertura-compatible)
	PyCoverXMLFormat
	// PyCoverJSONFormat indicates Python coverage JSON format
	PyCoverJSONFormat
)

// String returns the string representation of the coverage format
func (f CoverageFormat) String() string {
	switch f {
	case LCOVFormat:
		return "LCOV"
	case GoCoverFormat:
		return "Go Coverage"
	case PyCoverXMLFormat:
		return "Python XML Coverage"
	case PyCoverJSONFormat:
		return "Python JSON Coverage"
	default:
		return "Unknown"
	}
}

// DetectFormat attempts to detect the coverage file format by examining the file content
func DetectFormat(reader io.Reader) (CoverageFormat, error) {
	scanner := bufio.NewScanner(reader)

	// Read first few lines to determine format
	lineCount := 0
	maxLinesToCheck := 10

	hasLCOVMarkers := false
	hasGoMarkers := false
	hasXMLMarkers := false
	hasJSONMarkers := false

	for scanner.Scan() && lineCount < maxLinesToCheck {
		line := strings.TrimSpace(scanner.Text())

		lineCount++

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for Go coverage format markers
		// Go coverage files start with "mode: set|count|atomic"
		if lineCount == 1 && strings.HasPrefix(line, "mode:") {
			parts := strings.Fields(line)
			if len(parts) == 2 {
				mode := parts[1]
				if mode == "set" || mode == "count" || mode == "atomic" {
					hasGoMarkers = true
				}
			}
		}

		// Check for XML coverage format markers
		if strings.Contains(line, "<coverage") || strings.Contains(line, "<class filename=") {
			hasXMLMarkers = true
		}

		// Check for JSON coverage format markers
		if strings.Contains(line, `"files"`) || strings.Contains(line, `"executed_lines"`) {
			hasJSONMarkers = true
		}

		// Check for LCOV format markers
		if strings.HasPrefix(line, "TN:") ||
			strings.HasPrefix(line, "SF:") ||
			strings.HasPrefix(line, "FN:") ||
			strings.HasPrefix(line, "FNDA:") ||
			strings.HasPrefix(line, "DA:") ||
			strings.HasPrefix(line, "LH:") ||
			strings.HasPrefix(line, "LF:") ||
			line == "end_of_record" {
			hasLCOVMarkers = true
		}
	}

	if err := scanner.Err(); err != nil {
		return UnknownFormat, err
	}

	// Determine format based on markers found
	if hasGoMarkers {
		return GoCoverFormat, nil
	}

	if hasXMLMarkers {
		return PyCoverXMLFormat, nil
	}

	if hasJSONMarkers {
		return PyCoverJSONFormat, nil
	}

	if hasLCOVMarkers {
		return LCOVFormat, nil
	}

	return UnknownFormat, nil
}

// DetectFormatByExtension attempts to detect format based on file extension
func DetectFormatByExtension(filename string) CoverageFormat {
	filename = strings.ToLower(filename)

	// Go coverage files
	if strings.HasSuffix(filename, ".out") {
		return GoCoverFormat
	}

	// Python XML coverage files
	if strings.HasSuffix(filename, ".xml") && strings.Contains(filename, "coverage") {
		return PyCoverXMLFormat
	}

	// Python JSON coverage files
	if strings.HasSuffix(filename, ".json") && strings.Contains(filename, "coverage") {
		return PyCoverJSONFormat
	}

	// LCOV format files
	if strings.HasSuffix(filename, ".lcov") ||
		strings.HasSuffix(filename, ".info") ||
		strings.Contains(filename, "lcov.info") {
		return LCOVFormat
	}

	return UnknownFormat
}
