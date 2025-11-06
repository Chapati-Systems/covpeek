package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"git.kernel.fun/chapati.systems/covpeek/internal/detector"
	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
	"git.kernel.fun/chapati.systems/covpeek/pkg/parser"
	"github.com/spf13/cobra"
)

var minCoverage float64

var ciCmd = &cobra.Command{
	Use:   "ci --min <percentage>",
	Short: "Check total coverage against minimum threshold for CI",
	Long: `Automatically detect coverage files in standard locations, 
calculate total coverage, and fail if below the minimum threshold.`,
	Example: `  covpeek ci --min 80`,
	RunE:    runCI,
}

func init() {
	ciCmd.Flags().Float64Var(&minCoverage, "min", 0, "Minimum coverage percentage required (0-100)")
	ciCmd.MarkFlagRequired("min")
	rootCmd.AddCommand(ciCmd)
}

func runCI(cmd *cobra.Command, args []string) error {
	if minCoverage < 0 || minCoverage > 100 {
		return fmt.Errorf("--min must be between 0 and 100, got: %.2f", minCoverage)
	}

	existingFiles := detectExistingCoverageFiles()

	var reports []*models.CoverageReport

	for _, file := range existingFiles {
		// File exists, try to parse
		report, err := parseCoverageFile(file)
		if err != nil {
			cmd.PrintErrf("Warning: failed to parse %s: %v\n", file, err)
			continue
		}
		reports = append(reports, report)
	}

	if len(reports) == 0 {
		fmt.Println("No coverage files detected in standard locations. Please specify manually.")
		os.Exit(1)
	}

	// Merge reports
	mergedReport := mergeReports(reports)

	// Calculate overall coverage
	_, _, overallPct := mergedReport.CalculateOverallCoverage()

	// Check against threshold
	if overallPct >= minCoverage {
		fmt.Printf("Coverage check passed: %.2f%% >= %.0f%% threshold.\n", overallPct, minCoverage)
		os.Exit(0)
	} else {
		fmt.Printf("Coverage check failed: %.2f%% < %.0f%% minimum required.\n", overallPct, minCoverage)
		os.Exit(1)
	}

	return nil
}

func detectExistingCoverageFiles() []string {
	possibleFiles := getPossibleCoverageFiles()
	var existingFiles []string
	for _, file := range possibleFiles {
		if _, err := os.Stat(file); err == nil {
			existingFiles = append(existingFiles, file)
		}
	}
	return existingFiles
}

func getPossibleCoverageFiles() []string {
	return []string{
		"coverage.out",                 // Go
		"test/coverage.out",            // Go
		"lcov.info",                    // Rust/TS
		"target/coverage/lcov.info",    // Rust
		"coverage/lcov.info",           // TS
		"coverage/coverage-final.json", // TS
		"coverage.xml",                 // Python
		"coverage.json",                // Python
	}
}

func parseCoverageFile(filePath string) (*models.CoverageReport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Detect format
	format := detector.DetectFormatByExtension(filePath)
	if format == detector.UnknownFormat {
		var err error
		format, err = detector.DetectFormat(bytes.NewReader(content))
		if err != nil {
			return nil, err
		}
	}

	if format == detector.UnknownFormat {
		return nil, fmt.Errorf("unable to detect format")
	}

	// Parse
	var report *models.CoverageReport
	switch format {
	case detector.LCOVFormat:
		p := parser.NewLCOVParser()
		report, err = p.Parse(bytes.NewReader(content))
	case detector.GoCoverFormat:
		p := parser.NewGoCoverParser()
		report, err = p.Parse(bytes.NewReader(content))
	case detector.PyCoverXMLFormat:
		p := parser.NewPyCoverXMLParser()
		report, err = p.Parse(bytes.NewReader(content))
	case detector.PyCoverJSONFormat:
		p := parser.NewPyCoverJSONParser()
		report, err = p.Parse(bytes.NewReader(content))
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return report, err
}

func mergeReports(reports []*models.CoverageReport) *models.CoverageReport {
	merged := models.NewCoverageReport()
	for _, report := range reports {
		for _, file := range report.Files {
			// If file already exists, combine
			if existing := merged.GetFile(file.FileName); existing != nil {
				existing.TotalLines += file.TotalLines
				existing.CoveredLines += file.CoveredLines
				existing.CalculateCoverage()
				// Combine functions and lines if needed, but for simplicity, skip
			} else {
				merged.AddFile(file)
			}
		}
	}
	return merged
}
