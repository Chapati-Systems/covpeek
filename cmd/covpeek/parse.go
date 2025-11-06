package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Chapati-Systems/covpeek/internal/detector"
	"github.com/Chapati-Systems/covpeek/pkg/models"
	"github.com/Chapati-Systems/covpeek/pkg/parser"
	tea "github.com/charmbracelet/bubbletea"
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
	tuiMode      bool
)

func init() {
	// Define flags on root command
	rootCmd.Flags().StringVarP(&coverageFile, "file", "f", "", "Path to coverage file")
	rootCmd.Flags().StringVar(&forceFormat, "format", "", "Override format detection (rust, go, ts)")
	rootCmd.Flags().Float64VarP(&belowPct, "below", "b", 0, "Coverage threshold filter (0-100)")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, csv)")
	rootCmd.Flags().BoolVar(&tuiMode, "tui", false, "Launch interactive TUI for exploring coverage data")

	// Set the run function for the root command
	rootCmd.RunE = runParse
}

// validateFlags validates all flag inputs before execution
func validateFlags(cmd *cobra.Command, args []string) error {
	// Validate format if specified
	if forceFormat != "" {
		validFormats := map[string]bool{
			"rust":   true,
			"go":     true,
			"ts":     true,
			"lcov":   true, // alias for rust/ts
			"python": true,
			"pyxml":  true,
			"pyjson": true,
		}
		if !validFormats[strings.ToLower(forceFormat)] {
			return fmt.Errorf("invalid format '%s': must be one of: rust, go, ts, python, pyxml, pyjson", forceFormat)
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

	// If no file specified, auto-detect
	if coverageFile == "" {
		existingFiles := detectExistingCoverageFiles()
		if len(existingFiles) == 0 {
			return fmt.Errorf("no coverage files detected in standard locations. Please specify --file")
		}
		if len(existingFiles) > 1 {
			return fmt.Errorf("multiple coverage files detected: %v. Please specify --file", existingFiles)
		}
		coverageFile = existingFiles[0]
		cmd.PrintErrf("Auto-detected coverage file: %s\n", coverageFile)
	}

	// Validate file exists and is readable
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
	defer func() { _ = file.Close() }()

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
		case "python", "pyxml":
			format = detector.PyCoverXMLFormat
		case "pyjson":
			format = detector.PyCoverJSONFormat
		default:
			return fmt.Errorf("unknown format: %s (use 'rust', 'go', 'ts', 'python', 'pyxml', or 'pyjson')", forceFormat)
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

	case detector.PyCoverXMLFormat:
		p := parser.NewPyCoverXMLParser()
		report, err = p.Parse(bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("failed to parse Python XML coverage file: %w", err)
		}

	case detector.PyCoverJSONFormat:
		p := parser.NewPyCoverJSONParser()
		report, err = p.Parse(bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("failed to parse Python JSON coverage file: %w", err)
		}

	default:
		return fmt.Errorf("unsupported coverage format: %s", format)
	}

	// Apply threshold filter if specified
	if belowPct > 0 {
		report = filterBelowThreshold(report, belowPct)
	}

	// Output results
	if tuiMode {
		return outputTUI(report)
	}

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

// outputTUI launches an interactive TUI for exploring coverage data
func outputTUI(report *models.CoverageReport) error {
	// Create initial table model
	model := newTableModel(report)

	// Run the TUI with mouse support
	p := tea.NewProgram(model, tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
