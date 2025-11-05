package main

import (
	"git.kernel.fun/chapati.systems/covpeek/pkg/models"
)

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
