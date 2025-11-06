package main

import (
	"fmt"
	"os"

	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
	"github.com/spf13/cobra"
)

var (
	badgeFile   string
	badgeOutput string
	badgeLabel  string
	badgeStyle  string
)

var badgeCmd = &cobra.Command{
	Use:   "badge",
	Short: "Generate an SVG badge displaying total code coverage percentage",
	Long: `Automatically detect coverage files or use specified file, 
calculate total coverage, and generate an SVG badge similar to Shields.io.`,
	Example: `  covpeek badge --file coverage.lcov --output mybadge.svg
  covpeek badge --label "test coverage" --style plastic`,
	RunE: runBadge,
}

func init() {
	badgeCmd.Flags().StringVar(&badgeFile, "file", "", "Path to the coverage report file (optional, auto-detect if not provided)")
	badgeCmd.Flags().StringVar(&badgeOutput, "output", "coverage-badge.svg", "Path to save the generated SVG file")
	badgeCmd.Flags().StringVar(&badgeLabel, "label", "coverage", "Custom text label for the badge")
	badgeCmd.Flags().StringVar(&badgeStyle, "style", "flat", "Badge style: flat, plastic, flat-square")

	rootCmd.AddCommand(badgeCmd)
}

func runBadge(cmd *cobra.Command, args []string) error {
	// Validate style
	if badgeStyle != "flat" && badgeStyle != "plastic" && badgeStyle != "flat-square" {
		return fmt.Errorf("--style must be one of: flat, plastic, flat-square")
	}

	// Detect or parse coverage file
	var reports []*models.CoverageReport
	if badgeFile != "" {
		report, err := parseCoverageFile(badgeFile)
		if err != nil {
			return fmt.Errorf("failed to parse coverage file %s: %v", badgeFile, err)
		}
		reports = append(reports, report)
	} else {
		existingFiles := detectExistingCoverageFiles()
		if len(existingFiles) == 0 {
			return fmt.Errorf("no coverage files detected in standard locations. Please specify --file")
		}
		for _, file := range existingFiles {
			report, err := parseCoverageFile(file)
			if err != nil {
				cmd.PrintErrf("Warning: failed to parse %s: %v\n", file, err)
				continue
			}
			reports = append(reports, report)
		}
		if len(reports) == 0 {
			return fmt.Errorf("no valid coverage files found")
		}
	}

	// Merge reports
	mergedReport := mergeReports(reports)

	// Calculate overall coverage
	_, _, overallPct := mergedReport.CalculateOverallCoverage()

	// Generate badge
	color := getColorForCoverage(overallPct)
	svg := generateBadgeSVG(badgeLabel, fmt.Sprintf("%.1f%%", overallPct), color, badgeStyle)

	// Write to file
	err := os.WriteFile(badgeOutput, []byte(svg), 0644)
	if err != nil {
		return fmt.Errorf("failed to write SVG file: %v", err)
	}

	fmt.Printf("Badge generated: %s\n", badgeOutput)
	return nil
}

func getColorForCoverage(pct float64) string {
	switch {
	case pct >= 90:
		return "#4c1"
	case pct >= 80:
		return "#a4a61d"
	case pct >= 70:
		return "#dfb317"
	case pct >= 60:
		return "#fe7d37"
	default:
		return "#e05d44"
	}
}

func generateBadgeSVG(label, value, color, style string) string {
	labelWidth := len(label)*7 + 10 // rough estimate
	valueWidth := len(value)*7 + 10
	totalWidth := labelWidth + valueWidth

	rx := "3"
	gradient := ""
	overlay := ""
	if style == "flat-square" {
		rx = "0"
	} else if style == "plastic" {
		rx = "4"
		gradient = fmt.Sprintf(`<linearGradient id="a" x2="0" y2="100%%">
    <stop offset="0" stop-color="#fff" stop-opacity=".7"/>
    <stop offset=".1" stop-color="#aaa" stop-opacity=".1"/>
    <stop offset=".9" stop-opacity=".3"/>
    <stop offset="1" stop-opacity=".5"/>
  </linearGradient>`)
		overlay = fmt.Sprintf(`<rect width="%d" height="20" fill="url(#a)" rx="%s"/>`, totalWidth, rx)
	}

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20">
  %s
  <rect width="%d" height="20" fill="#555" rx="%s"/>
  <rect x="%d" width="%d" height="20" fill="%s" rx="%s"/>
  %s
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`, totalWidth, gradient, labelWidth, rx, labelWidth, valueWidth, color, rx, overlay, labelWidth/2, label, labelWidth/2, label, labelWidth+valueWidth/2, value, labelWidth+valueWidth/2, value)

	return svg
}
