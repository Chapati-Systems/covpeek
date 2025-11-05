package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"git.kernel.fun/chapati.systems/covpeek/internal/detector"
	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
	"git.kernel.fun/chapati.systems/covpeek/pkg/parser"
	"github.com/spf13/cobra"
)

// Config holds validated configuration from CLI flags
type Config struct {
	FilePath     string
	OutputFormat string
	ForceFormat  string
	BelowPct     float64
}

var (
	coverageFile string
	outputFormat string
	forceFormat  string
	belowPct     float64
)

func init() {
	// Define flags on root command
	rootCmd.Flags().StringVarP(&coverageFile, "file", "f", "", "Path to coverage file (required)")
	rootCmd.Flags().StringVar(&forceFormat, "format", "", "Override format detection (rust, go, ts)")
	rootCmd.Flags().Float64Var(&belowPct, "below", 0, "Coverage threshold filter (0-100)")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, csv)")

	// Set the run function for the root command
	rootCmd.RunE = runParse
}

// validateFlags validates all flag inputs before execution
func validateFlags(cmd *cobra.Command, args []string) error {
	// Validate file exists and is readable
	if coverageFile == "" {
		return fmt.Errorf("--file flag is required")
	}

	fileInfo, err := os.Stat(coverageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", coverageFile)
		}
		return fmt.Errorf("cannot access file %s: %w", coverageFile, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", coverageFile)
	}

	// Test if file is readable
	file, err := os.Open(coverageFile)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", coverageFile, err)
	}
	file.Close()

	// Validate format if specified
	if forceFormat != "" {
		validFormats := map[string]bool{
			"rust": true,
			"go":   true,
			"ts":   true,
			"lcov": true, // alias for rust/ts
		}
		if !validFormats[strings.ToLower(forceFormat)] {
			return fmt.Errorf("invalid format '%s': must be one of: rust, go, ts", forceFormat)
		}
	}

	// Validate below percentage
	if belowPct < 0 || belowPct > 100 {
		return fmt.Errorf("--below must be between 0 and 100, got: %.2f", belowPct)
	}

	// Validate output format
	validOutputs := map[string]bool{
		"table": true,
		"json":  true,
		"csv":   true,
	}
	if !validOutputs[strings.ToLower(outputFormat)] {
		return fmt.Errorf("invalid output format '%s': must be one of: table, json, csv", outputFormat)
	}

	return nil
}

func runParse(cmd *cobra.Command, args []string) error {
	// If "help" is passed as an argument, show help instead of trying to parse
	if len(args) > 0 && args[0] == "help" {
		return cmd.Help()
	}

	// Validate flags first
	if err := validateFlags(cmd, args); err != nil {
		return err
	}

	// Open coverage file
	file, err := os.Open(coverageFile)
	if err != nil {
		return fmt.Errorf("failed to open coverage file: %w", err)
	}
	defer file.Close()

	// Read file content into memory for detection and parsing
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read coverage file: %w", err)
	}

	// Detect format
	var format detector.CoverageFormat
	if forceFormat != "" {
		// Map format names to detector format
		switch strings.ToLower(forceFormat) {
		case "lcov", "rust", "ts":
			format = detector.LCOVFormat
		case "go":
			format = detector.GoCoverFormat
		default:
			return fmt.Errorf("unknown format: %s (use 'rust', 'go', or 'ts')", forceFormat)
		}
		cmd.PrintErrf("Using forced format: %s\n", format)
	} else {
		// Try detection by extension first
		format = detector.DetectFormatByExtension(coverageFile)
		if format == detector.UnknownFormat {
			// Fall back to content-based detection
			format, err = detector.DetectFormat(bytes.NewReader(content))
			if err != nil {
				return fmt.Errorf("failed to detect coverage format: %w", err)
			}
		}

		if format == detector.UnknownFormat {
			return fmt.Errorf("unable to detect coverage format for file: %s", coverageFile)
		}

		cmd.PrintErrf("Detected format: %s\n", format)
	}

	// Parse based on detected format
	var report *models.CoverageReport
	switch format {
	case detector.LCOVFormat:
		p := parser.NewLCOVParser()
		report, err = p.Parse(bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("failed to parse LCOV file: %w", err)
		}

	case detector.GoCoverFormat:
		p := parser.NewGoCoverParser()
		report, err = p.Parse(bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("failed to parse Go coverage file: %w", err)
		}

	default:
		return fmt.Errorf("unsupported coverage format: %s", format)
	}

	// Apply threshold filter if specified
	if belowPct > 0 {
		report = filterBelowThreshold(report, belowPct)
	}

	// Output results
	switch strings.ToLower(outputFormat) {
	case "json":
		return outputJSON(report)
	case "csv":
		return outputCSV(report)
	case "table":
		return outputTable(report)
	default:
		return outputTable(report)
	}
}

// filterBelowThreshold filters files with coverage below the threshold
func filterBelowThreshold(report *models.CoverageReport, threshold float64) *models.CoverageReport {
	filtered := models.NewCoverageReport()
	filtered.TestName = report.TestName

	for _, fileCov := range report.Files {
		if fileCov.CoveragePct < threshold {
			filtered.AddFile(fileCov)
		}
	}

	return filtered
}

// outputTable outputs coverage data in a readable table format
func outputTable(report *models.CoverageReport) error {
	if report.TestName != "" {
		fmt.Printf("Test Name: %s\n\n", report.TestName)
	}

	if len(report.Files) == 0 {
		fmt.Println("No files found in coverage report")
		return nil
	}

	// Calculate overall coverage
	totalLines := 0
	totalCovered := 0
	for _, fileCov := range report.Files {
		totalLines += fileCov.TotalLines
		totalCovered += fileCov.CoveredLines
	}

	// Sort files by name for consistent output
	filenames := make([]string, 0, len(report.Files))
	for filename := range report.Files {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	// Print table header
	fmt.Println("┌────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                         Coverage Report                                   │")
	fmt.Println("├────────────────────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ %-50s │ %10s │ %10s │\n", "File", "Coverage", "Lines")
	fmt.Println("├────────────────────────────────────────────────────────────────────────────┤")

	// Print file rows
	for _, filename := range filenames {
		fileCov := report.Files[filename]
		// Truncate filename if too long
		displayName := filename
		if len(displayName) > 50 {
			displayName = "..." + displayName[len(displayName)-47:]
		}
		fmt.Printf("│ %-50s │ %9.2f%% │ %4d/%4d │\n",
			displayName,
			fileCov.CoveragePct,
			fileCov.CoveredLines,
			fileCov.TotalLines)
	}

	// Print summary
	fmt.Println("├────────────────────────────────────────────────────────────────────────────┤")
	overallPct := 0.0
	if totalLines > 0 {
		overallPct = (float64(totalCovered) / float64(totalLines)) * 100.0
	}
	fmt.Printf("│ %-50s │ %9.2f%% │ %4d/%4d │\n",
		"OVERALL",
		overallPct,
		totalCovered,
		totalLines)
	fmt.Println("└────────────────────────────────────────────────────────────────────────────┘")

	return nil
}

// outputJSON outputs coverage data in JSON format
func outputJSON(report *models.CoverageReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// outputCSV outputs coverage data in CSV format
func outputCSV(report *models.CoverageReport) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"File", "Coverage %", "Covered Lines", "Total Lines"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Sort files by name for consistent output
	filenames := make([]string, 0, len(report.Files))
	for filename := range report.Files {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	// Write file rows
	for _, filename := range filenames {
		fileCov := report.Files[filename]
		row := []string{
			filename,
			fmt.Sprintf("%.2f", fileCov.CoveragePct),
			fmt.Sprintf("%d", fileCov.CoveredLines),
			fmt.Sprintf("%d", fileCov.TotalLines),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
