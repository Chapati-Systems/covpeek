package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
)

// GoCoverParser parses Go coverage format (.out) files
type GoCoverParser struct {
	warnings []string
	mode     string
}

// NewGoCoverParser creates a new Go coverage parser instance
func NewGoCoverParser() *GoCoverParser {
	return &GoCoverParser{
		warnings: make([]string, 0),
	}
}

// Parse reads and parses a Go coverage format file
// Format: mode: set|count|atomic
// Then: file:startLine.startCol,endLine.endCol numberOfStatements count
func (p *GoCoverParser) Parse(reader io.Reader) (*models.CoverageReport, error) {
	report := models.NewCoverageReport()
	scanner := bufio.NewScanner(reader)
	lineNumber := 0

	// First line should be mode declaration
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty coverage file")
	}

	lineNumber++
	firstLine := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(firstLine, "mode:") {
		return nil, fmt.Errorf("invalid Go coverage file: missing mode declaration on line 1")
	}

	p.mode = strings.TrimSpace(strings.TrimPrefix(firstLine, "mode:"))
	if p.mode != "set" && p.mode != "count" && p.mode != "atomic" {
		p.addWarning(lineNumber, fmt.Sprintf("unknown coverage mode: %s", p.mode))
	}

	// Parse coverage entries
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		if err := p.parseCoverageEntry(line, report, lineNumber); err != nil {
			p.addWarning(lineNumber, err.Error())
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading coverage file: %w", err)
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

// parseCoverageEntry parses a single coverage entry line
// Format: file:startLine.startCol,endLine.endCol numberOfStatements count
func (p *GoCoverParser) parseCoverageEntry(line string, report *models.CoverageReport, lineNum int) error {
	// Split by colon to separate filename from coverage data
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return fmt.Errorf("invalid coverage entry format: %s", line)
	}

	filename := line[:colonIdx]
	coverageData := line[colonIdx+1:]

	// Parse coverage data: startLine.startCol,endLine.endCol numberOfStatements count
	parts := strings.Fields(coverageData)
	if len(parts) != 3 {
		return fmt.Errorf("invalid coverage data format: %s", coverageData)
	}

	// Parse line range: startLine.startCol,endLine.endCol
	lineRange := parts[0]
	rangeParts := strings.Split(lineRange, ",")
	if len(rangeParts) != 2 {
		return fmt.Errorf("invalid line range format: %s", lineRange)
	}

	startParts := strings.Split(rangeParts[0], ".")
	if len(startParts) != 2 {
		return fmt.Errorf("invalid start position format: %s", rangeParts[0])
	}

	endParts := strings.Split(rangeParts[1], ".")
	if len(endParts) != 2 {
		return fmt.Errorf("invalid end position format: %s", rangeParts[1])
	}

	startLine, err := strconv.Atoi(startParts[0])
	if err != nil {
		return fmt.Errorf("invalid start line number: %s", startParts[0])
	}

	endLine, err := strconv.Atoi(endParts[0])
	if err != nil {
		return fmt.Errorf("invalid end line number: %s", endParts[0])
	}

	// Parse number of statements
	numStatements, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid number of statements: %s", parts[1])
	}

	// Parse execution count
	execCount, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid execution count: %s", parts[2])
	}

	// Get or create file coverage entry
	file := report.GetFile(filename)
	if file == nil {
		file = &models.FileCoverage{
			FileName:  filename,
			Functions: make([]models.FunctionCoverage, 0),
			Lines:     make(map[int]models.LineCoverage),
		}
		report.AddFile(file)
	}

	// Add line coverage data for each line in the range
	for lineNo := startLine; lineNo <= endLine; lineNo++ {
		// If line already exists, update count (take max or sum based on mode)
		if existing, exists := file.Lines[lineNo]; exists {
			// For count mode, we sum the counts; for set mode, we just mark as covered
			if p.mode == "count" || p.mode == "atomic" {
				existing.ExecutionCount += execCount
			} else {
				if execCount > 0 {
					existing.ExecutionCount = 1
				}
			}
			file.Lines[lineNo] = existing
		} else {
			file.Lines[lineNo] = models.LineCoverage{
				LineNumber:     lineNo,
				ExecutionCount: execCount,
			}
		}
	}

	// Update total lines and covered lines
	for _, lineCov := range file.Lines {
		if lineCov.ExecutionCount > 0 {
			// Count covered lines
			found := false
			for existingLine := range file.Lines {
				if existingLine == lineCov.LineNumber {
					found = true
					break
				}
			}
			if !found {
				file.CoveredLines++
			}
		}
	}

	// Update total lines count
	file.TotalLines = len(file.Lines)
	
	// Count covered lines
	coveredCount := 0
	for _, lineCov := range file.Lines {
		if lineCov.ExecutionCount > 0 {
			coveredCount++
		}
	}
	file.CoveredLines = coveredCount

	// Silently ignore numStatements for now (could be used for validation)
	_ = numStatements

	return nil
}

// addWarning adds a warning message to the parser
func (p *GoCoverParser) addWarning(lineNum int, message string) {
	p.warnings = append(p.warnings, fmt.Sprintf("line %d: %s", lineNum, message))
}

// GetWarnings returns all warnings collected during parsing
func (p *GoCoverParser) GetWarnings() []string {
	return p.warnings
}

// GetMode returns the coverage mode (set, count, or atomic)
func (p *GoCoverParser) GetMode() string {
	return p.mode
}
