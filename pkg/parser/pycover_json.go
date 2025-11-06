package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Chapati-Systems/covpeek/pkg/models"
)

// PyCoverJSONParser parses Python coverage JSON format
type PyCoverJSONParser struct {
	warnings []string
}

// NewPyCoverJSONParser creates a new Python JSON coverage parser instance
func NewPyCoverJSONParser() *PyCoverJSONParser {
	return &PyCoverJSONParser{
		warnings: make([]string, 0),
	}
}

// CoverageJSON represents the structure of Python coverage JSON
type CoverageJSON struct {
	Files map[string]FileCoverageJSON `json:"files"`
}

// FileCoverageJSON represents coverage data for a single file in JSON
type FileCoverageJSON struct {
	ExecutedLines []int       `json:"executed_lines"`
	MissingLines  []int       `json:"missing_lines"`
	Summary       SummaryJSON `json:"summary"`
}

// SummaryJSON represents the summary data
type SummaryJSON struct {
	CoveredLines  int `json:"covered_lines"`
	NumStatements int `json:"num_statements"`
}

// Parse reads and parses a Python coverage JSON file
func (p *PyCoverJSONParser) Parse(reader io.Reader) (*models.CoverageReport, error) {
	report := models.NewCoverageReport()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON data: %w", err)
	}

	var coverageJSON CoverageJSON
	if err := json.Unmarshal(data, &coverageJSON); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Process each file
	for filename, fileData := range coverageJSON.Files {
		if err := p.parseFile(filename, fileData, report); err != nil {
			p.addWarning(fmt.Sprintf("failed to parse file %s: %v", filename, err))
			continue
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

// parseFile parses a single file from the JSON
func (p *PyCoverJSONParser) parseFile(filename string, fileData FileCoverageJSON, report *models.CoverageReport) error {
	// Normalize filename (remove leading ./ if present)
	filename = strings.TrimPrefix(filename, "./")

	file := &models.FileCoverage{
		FileName:  filename,
		Functions: make([]models.FunctionCoverage, 0),
		Lines:     make(map[int]models.LineCoverage),
	}

	// Process executed lines
	for _, lineNum := range fileData.ExecutedLines {
		file.Lines[lineNum] = models.LineCoverage{
			LineNumber:     lineNum,
			ExecutionCount: 1, // JSON format doesn't specify hit counts, just executed/missing
		}
	}

	// Process missing lines (ensure they're in the map with 0 hits)
	for _, lineNum := range fileData.MissingLines {
		if _, exists := file.Lines[lineNum]; !exists {
			file.Lines[lineNum] = models.LineCoverage{
				LineNumber:     lineNum,
				ExecutionCount: 0,
			}
		}
	}

	// Use summary data if available
	if fileData.Summary.NumStatements > 0 {
		file.TotalLines = fileData.Summary.NumStatements
		file.CoveredLines = fileData.Summary.CoveredLines

		// Validate summary data against line data
		lineTotal := len(file.Lines)
		lineCovered := 0
		for _, lineCov := range file.Lines {
			if lineCov.ExecutionCount > 0 {
				lineCovered++
			}
		}

		if fileData.Summary.NumStatements != lineTotal {
			p.addWarning(fmt.Sprintf("file %s: summary num_statements (%d) doesn't match line count (%d)", filename, fileData.Summary.NumStatements, lineTotal))
		}
		if fileData.Summary.CoveredLines != lineCovered {
			p.addWarning(fmt.Sprintf("file %s: summary covered_lines (%d) doesn't match calculated covered lines (%d)", filename, fileData.Summary.CoveredLines, lineCovered))
		}
	} else {
		// Fallback: calculate from line data
		file.TotalLines = len(file.Lines)
		coveredCount := 0
		for _, lineCov := range file.Lines {
			if lineCov.ExecutionCount > 0 {
				coveredCount++
			}
		}
		file.CoveredLines = coveredCount
	}

	report.AddFile(file)
	return nil
}

// addWarning adds a warning message to the parser
func (p *PyCoverJSONParser) addWarning(message string) {
	p.warnings = append(p.warnings, message)
}

// GetWarnings returns all warnings collected during parsing
func (p *PyCoverJSONParser) GetWarnings() []string {
	return p.warnings
}
