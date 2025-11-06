package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/Chapati-Systems/covpeek/pkg/models"
)

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
	totalLines, totalCovered, overallPct := report.CalculateOverallCoverage()

	// Sort files by coverage descending (highest first)
	type fileEntry struct {
		name string
		cov  *models.FileCoverage
	}

	entries := make([]fileEntry, 0, len(report.Files))
	for name, cov := range report.Files {
		entries = append(entries, fileEntry{name: name, cov: cov})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].cov.CoveragePct > entries[j].cov.CoveragePct
	})

	// Create tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		}
	}()

	// Print header
	if _, err := fmt.Fprintln(w, "File\tTotal Lines\tCovered Lines\tCoverage %"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "----\t----------\t-------------\t----------"); err != nil {
		return fmt.Errorf("failed to write table separator: %w", err)
	}

	// Print file rows
	for _, entry := range entries {
		if _, err := fmt.Fprintf(w, "%s\t%d\t%d\t%.2f%%\n",
			entry.name,
			entry.cov.TotalLines,
			entry.cov.CoveredLines,
			entry.cov.CoveragePct); err != nil {
			return fmt.Errorf("failed to write table row: %w", err)
		}
	}

	// Print overall summary
	if _, err := fmt.Fprintf(w, "\nOverall\t%d\t%d\t%.2f%%\n",
		totalLines,
		totalCovered,
		overallPct); err != nil {
		return fmt.Errorf("failed to write table summary: %w", err)
	}

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
