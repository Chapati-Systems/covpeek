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

// LCOVParser parses LCOV format coverage files
type LCOVParser struct {
	warnings []string
}

// NewLCOVParser creates a new LCOV parser instance
func NewLCOVParser() *LCOVParser {
	return &LCOVParser{
		warnings: make([]string, 0),
	}
}

// Parse reads and parses an LCOV format coverage file
func (p *LCOVParser) Parse(reader io.Reader) (*models.CoverageReport, error) {
	report := models.NewCoverageReport()
	scanner := bufio.NewScanner(reader)

	var currentFile *models.FileCoverage
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse different LCOV record types
		switch {
		case strings.HasPrefix(line, "TN:"):
			// Test name
			report.TestName = strings.TrimPrefix(line, "TN:")

		case strings.HasPrefix(line, "SF:"):
			// Source file
			filename := strings.TrimPrefix(line, "SF:")
			currentFile = &models.FileCoverage{
				FileName:  filename,
				Functions: make([]models.FunctionCoverage, 0),
				Lines:     make(map[int]models.LineCoverage),
			}

		case strings.HasPrefix(line, "FN:"):
			// Function definition: FN:<line>,<function name>
			if currentFile == nil {
				p.addWarning(lineNumber, "FN record without active source file")
				continue
			}
			if err := p.parseFN(line, currentFile, lineNumber); err != nil {
				p.addWarning(lineNumber, err.Error())
			}

		case strings.HasPrefix(line, "FNDA:"):
			// Function data: FNDA:<execution count>,<function name>
			if currentFile == nil {
				p.addWarning(lineNumber, "FNDA record without active source file")
				continue
			}
			if err := p.parseFNDA(line, currentFile, lineNumber); err != nil {
				p.addWarning(lineNumber, err.Error())
			}

		case strings.HasPrefix(line, "FNF:"):
			// Functions found (total count) - we can validate this
			// FNF:<number of functions found>
			continue

		case strings.HasPrefix(line, "FNH:"):
			// Functions hit (executed) - we can validate this
			// FNH:<number of functions hit>
			continue

		case strings.HasPrefix(line, "DA:"):
			// Line data: DA:<line number>,<execution count>[,<checksum>]
			if currentFile == nil {
				p.addWarning(lineNumber, "DA record without active source file")
				continue
			}
			if err := p.parseDA(line, currentFile, lineNumber); err != nil {
				p.addWarning(lineNumber, err.Error())
			}

		case strings.HasPrefix(line, "LH:"):
			// Lines hit: LH:<number of lines with non-zero execution count>
			if currentFile == nil {
				p.addWarning(lineNumber, "LH record without active source file")
				continue
			}
			if err := p.parseLH(line, currentFile, lineNumber); err != nil {
				p.addWarning(lineNumber, err.Error())
			}

		case strings.HasPrefix(line, "LF:"):
			// Lines found: LF:<number of instrumented lines>
			if currentFile == nil {
				p.addWarning(lineNumber, "LF record without active source file")
				continue
			}
			if err := p.parseLF(line, currentFile, lineNumber); err != nil {
				p.addWarning(lineNumber, err.Error())
			}

		case strings.HasPrefix(line, "BRF:"):
			// Branches found - branch coverage (optional)
			continue

		case strings.HasPrefix(line, "BRH:"):
			// Branches hit - branch coverage (optional)
			continue

		case strings.HasPrefix(line, "BRDA:"):
			// Branch data - branch coverage (optional)
			continue

		case line == "end_of_record":
			// End of current file record
			if currentFile != nil {
				currentFile.CalculateCoverage()
				report.AddFile(currentFile)
				currentFile = nil
			}

		default:
			// Unknown record type - log warning but continue
			p.addWarning(lineNumber, fmt.Sprintf("unknown record type: %s", line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading coverage file: %w", err)
	}

	// Log all warnings
	for _, warning := range p.warnings {
		log.Println(warning)
	}

	return report, nil
}

// parseFN parses function definition: FN:<line>,<function name>
func (p *LCOVParser) parseFN(line string, file *models.FileCoverage, lineNum int) error {
	data := strings.TrimPrefix(line, "FN:")
	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid FN format: %s", line)
	}

	lineNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid line number in FN: %s", parts[0])
	}

	functionName := parts[1]
	file.Functions = append(file.Functions, models.FunctionCoverage{
		Name:       functionName,
		LineNumber: lineNumber,
	})

	return nil
}

// parseFNDA parses function data: FNDA:<execution count>,<function name>
func (p *LCOVParser) parseFNDA(line string, file *models.FileCoverage, lineNum int) error {
	data := strings.TrimPrefix(line, "FNDA:")
	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid FNDA format: %s", line)
	}

	execCount, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid execution count in FNDA: %s", parts[0])
	}

	functionName := parts[1]

	// Find the matching function and update execution count
	for i := range file.Functions {
		if file.Functions[i].Name == functionName {
			file.Functions[i].ExecutionCount = execCount
			break
		}
	}

	return nil
}

// parseDA parses line data: DA:<line number>,<execution count>[,<checksum>]
func (p *LCOVParser) parseDA(line string, file *models.FileCoverage, lineNum int) error {
	data := strings.TrimPrefix(line, "DA:")
	parts := strings.Split(data, ",")
	if len(parts) < 2 {
		return fmt.Errorf("invalid DA format: %s", line)
	}

	lineNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid line number in DA: %s", parts[0])
	}

	execCount, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid execution count in DA: %s", parts[1])
	}

	checksum := ""
	if len(parts) > 2 {
		checksum = parts[2]
	}

	file.Lines[lineNumber] = models.LineCoverage{
		LineNumber:     lineNumber,
		ExecutionCount: execCount,
		Checksum:       checksum,
	}

	return nil
}

// parseLH parses lines hit: LH:<number>
func (p *LCOVParser) parseLH(line string, file *models.FileCoverage, lineNum int) error {
	data := strings.TrimPrefix(line, "LH:")
	count, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("invalid LH value: %s", data)
	}

	file.CoveredLines = count
	return nil
}

// parseLF parses lines found: LF:<number>
func (p *LCOVParser) parseLF(line string, file *models.FileCoverage, lineNum int) error {
	data := strings.TrimPrefix(line, "LF:")
	count, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("invalid LF value: %s", data)
	}

	file.TotalLines = count
	return nil
}

// addWarning adds a warning message to the parser
func (p *LCOVParser) addWarning(lineNum int, message string) {
	p.warnings = append(p.warnings, fmt.Sprintf("line %d: %s", lineNum, message))
}

// GetWarnings returns all warnings collected during parsing
func (p *LCOVParser) GetWarnings() []string {
	return p.warnings
}
