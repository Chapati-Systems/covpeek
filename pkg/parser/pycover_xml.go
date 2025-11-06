package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"

	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
)

// PyCoverXMLParser parses Python coverage XML format (Cobertura-compatible)
type PyCoverXMLParser struct {
	warnings []string
}

// NewPyCoverXMLParser creates a new Python XML coverage parser instance
func NewPyCoverXMLParser() *PyCoverXMLParser {
	return &PyCoverXMLParser{
		warnings: make([]string, 0),
	}
}

// CoverageReport represents the root XML element
type CoverageReport struct {
	XMLName  xml.Name  `xml:"coverage"`
	Packages []Package `xml:"packages>package"`
}

// Package represents a package in the coverage report
type Package struct {
	Classes []Class `xml:"classes>class"`
}

// Class represents a class/file in the coverage report
type Class struct {
	Filename string `xml:"filename,attr"`
	Lines    []Line `xml:"lines>line"`
}

// Line represents a line in the coverage report
type Line struct {
	Number int `xml:"number,attr"`
	Hits   int `xml:"hits,attr"`
}

// Parse reads and parses a Python coverage XML file
func (p *PyCoverXMLParser) Parse(reader io.Reader) (*models.CoverageReport, error) {
	report := models.NewCoverageReport()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read XML data: %w", err)
	}

	var coverageReport CoverageReport
	if err := xml.Unmarshal(data, &coverageReport); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Process each package
	for _, pkg := range coverageReport.Packages {
		for _, class := range pkg.Classes {
			if err := p.parseClass(class, report); err != nil {
				p.addWarning(fmt.Sprintf("failed to parse class %s: %v", class.Filename, err))
				continue
			}
		}
	}

	// Calculate coverage for all files
	for _, file := range report.Files {
		file.CalculateCoverage()
	}

	// Log all warnings
	for _, warning := range p.warnings {
		log.Println(warning)
	}

	return report, nil
}

// parseClass parses a single class/file from the XML
func (p *PyCoverXMLParser) parseClass(class Class, report *models.CoverageReport) error {
	// Normalize filename (remove leading ./ if present)
	filename := strings.TrimPrefix(class.Filename, "./")

	file := &models.FileCoverage{
		FileName:  filename,
		Functions: make([]models.FunctionCoverage, 0),
		Lines:     make(map[int]models.LineCoverage),
	}

	// Process lines
	for _, line := range class.Lines {
		if line.Hits < 0 {
			p.addWarning(fmt.Sprintf("file %s: line %d has negative hit count %d", filename, line.Number, line.Hits))
		}
		file.Lines[line.Number] = models.LineCoverage{
			LineNumber:     line.Number,
			ExecutionCount: line.Hits,
		}
	}

	// Calculate total and covered lines
	file.TotalLines = len(file.Lines)
	coveredCount := 0
	for _, lineCov := range file.Lines {
		if lineCov.ExecutionCount > 0 {
			coveredCount++
		}
	}
	file.CoveredLines = coveredCount

	report.AddFile(file)
	return nil
}

// addWarning adds a warning message to the parser
func (p *PyCoverXMLParser) addWarning(message string) {
	p.warnings = append(p.warnings, message)
}

// GetWarnings returns all warnings collected during parsing
func (p *PyCoverXMLParser) GetWarnings() []string {
	return p.warnings
}
