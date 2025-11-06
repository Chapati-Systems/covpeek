package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"git.kernel.fun/chapati.systems/covpeek/internal/detector"
	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
	"git.kernel.fun/chapati.systems/covpeek/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	diffFile         string
	commitA          string
	commitB          string
	diffOutputFormat string
)

var diffCmd = &cobra.Command{
	Use:     "diff --file <path> [flags]",
	Short:   "Compare coverage reports between two git commits",
	Long:    `Compare coverage reports from two different git commits, showing changes in overall and per-file coverage.`,
	Example: `  covpeek diff --file coverage/lcov.info --commit-a HEAD~5 --commit-b HEAD`,
	RunE:    runDiff,
}

func init() {
	diffCmd.Flags().StringVarP(&diffFile, "file", "f", "", "Path to the coverage file relative to the repo root")
	diffCmd.Flags().StringVar(&commitA, "commit-a", "HEAD~1", "Git commit hash or ref for the base coverage report")
	diffCmd.Flags().StringVar(&commitB, "commit-b", "HEAD", "Git commit hash or ref for the target coverage report")
	diffCmd.Flags().StringVar(&diffOutputFormat, "output", "detailed", "Output format: summary, detailed, json")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	// Validate output format
	if diffOutputFormat != "summary" && diffOutputFormat != "detailed" && diffOutputFormat != "json" {
		return fmt.Errorf("invalid output format: %s. Must be summary, detailed, or json", diffOutputFormat)
	}

	// If no file specified, auto-detect
	if diffFile == "" {
		existingFiles := detectExistingCoverageFiles()
		if len(existingFiles) == 0 {
			return fmt.Errorf("no coverage files detected in standard locations. Please specify --file")
		}
		if len(existingFiles) > 1 {
			return fmt.Errorf("multiple coverage files detected: %v. Please specify --file", existingFiles)
		}
		diffFile = existingFiles[0]
		cmd.PrintErrf("Auto-detected coverage file: %s\n", diffFile)
	}

	// Get content from commitA
	contentA, err := getFileFromCommit(commitA, diffFile)
	if err != nil {
		return fmt.Errorf("failed to get coverage file from commit %s: %v", commitA, err)
	}

	// Get content from commitB
	contentB, err := getFileFromCommit(commitB, diffFile)
	if err != nil {
		return fmt.Errorf("failed to get coverage file from commit %s: %v", commitB, err)
	}

	// Parse reports
	reportA, err := parseCoverageContent(contentA, diffFile)
	if err != nil {
		return fmt.Errorf("failed to parse coverage from commit %s: %v", commitA, err)
	}

	reportB, err := parseCoverageContent(contentB, diffFile)
	if err != nil {
		return fmt.Errorf("failed to parse coverage from commit %s: %v", commitB, err)
	}

	// Compute diff
	diff := computeDiff(reportA, reportB)

	// Output
	outputDiff(diff, diffOutputFormat)

	return nil
}

func outputDiff(diff *CoverageDiff, format string) {
	switch format {
	case "summary":
		fmt.Printf("Overall coverage changed: %.1f%% -> %.1f%% (%.1f%%)\n", diff.OverallA, diff.OverallB, diff.OverallDelta)
	case "detailed":
		fmt.Printf("Overall coverage changed: %.1f%% -> %.1f%% (%.1f%%)\n", diff.OverallA, diff.OverallB, diff.OverallDelta)
		if len(diff.FileChanges) > 0 {
			fmt.Println("File coverage changes:")
			for _, fc := range diff.FileChanges {
				if fc.Delta == 0 {
					fmt.Printf("  %s: %.1f%% -> %.1f%% (no change)\n", fc.FileName, fc.CoverageA, fc.CoverageB)
				} else {
					sign := "+"
					if fc.Delta < 0 {
						sign = ""
					}
					fmt.Printf("  %s: %.1f%% -> %.1f%% (%s%.1f%%)\n", fc.FileName, fc.CoverageA, fc.CoverageB, sign, fc.Delta)
				}
			}
		}
	case "json":
		jsonData, _ := json.MarshalIndent(diff, "", "  ")
		fmt.Println(string(jsonData))
	}
}

func getFileFromCommit(commit, file string) ([]byte, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", commit, file))
	return cmd.Output()
}

func parseCoverageContent(content []byte, filePath string) (*models.CoverageReport, error) {
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
	var err error
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

// CoverageDiff represents the diff between two coverage reports
type CoverageDiff struct {
	OverallA     float64      `json:"overall_a"`
	OverallB     float64      `json:"overall_b"`
	OverallDelta float64      `json:"overall_delta"`
	FileChanges  []FileChange `json:"file_changes"`
}

// FileChange represents coverage change for a file
type FileChange struct {
	FileName  string  `json:"file_name"`
	CoverageA float64 `json:"coverage_a"`
	CoverageB float64 `json:"coverage_b"`
	Delta     float64 `json:"delta"`
}

func computeDiff(reportA, reportB *models.CoverageReport) *CoverageDiff {
	_, _, overallA := reportA.CalculateOverallCoverage()
	_, _, overallB := reportB.CalculateOverallCoverage()
	delta := overallB - overallA

	fileChanges := []FileChange{}

	// Collect all files
	fileMap := make(map[string]bool)
	for fname := range reportA.Files {
		fileMap[fname] = true
	}
	for fname := range reportB.Files {
		fileMap[fname] = true
	}

	for fname := range fileMap {
		fcA := reportA.GetFile(fname)
		fcB := reportB.GetFile(fname)

		covA := 0.0
		if fcA != nil {
			covA = fcA.CoveragePct
		}

		covB := 0.0
		if fcB != nil {
			covB = fcB.CoveragePct
		}

		d := covB - covA

		fileChanges = append(fileChanges, FileChange{
			FileName:  fname,
			CoverageA: covA,
			CoverageB: covB,
			Delta:     d,
		})
	}

	return &CoverageDiff{
		OverallA:     overallA,
		OverallB:     overallB,
		OverallDelta: delta,
		FileChanges:  fileChanges,
	}
}
