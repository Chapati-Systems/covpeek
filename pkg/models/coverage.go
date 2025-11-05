package models

// FileCoverage represents coverage data for a single source file
type FileCoverage struct {
	FileName     string
	TotalLines   int
	CoveredLines int
	CoveragePct  float64
	Functions    []FunctionCoverage
	Lines        map[int]LineCoverage
}

// FunctionCoverage represents coverage data for a function
type FunctionCoverage struct {
	Name           string
	LineNumber     int
	ExecutionCount int
}

// LineCoverage represents coverage data for a single line
type LineCoverage struct {
	LineNumber     int
	ExecutionCount int
	Checksum       string
}

// CoverageReport represents the complete coverage report
type CoverageReport struct {
	TestName string
	Files    map[string]*FileCoverage
}

// NewCoverageReport creates a new empty coverage report
func NewCoverageReport() *CoverageReport {
	return &CoverageReport{
		Files: make(map[string]*FileCoverage),
	}
}

// AddFile adds a file coverage entry to the report
func (r *CoverageReport) AddFile(fc *FileCoverage) {
	r.Files[fc.FileName] = fc
}

// GetFile retrieves file coverage by filename
func (r *CoverageReport) GetFile(filename string) *FileCoverage {
	return r.Files[filename]
}

// CalculateOverallCoverage calculates the total lines, covered lines, and overall coverage percentage across all files
func (r *CoverageReport) CalculateOverallCoverage() (totalLines int, totalCovered int, overallPct float64) {
	for _, fc := range r.Files {
		totalLines += fc.TotalLines
		totalCovered += fc.CoveredLines
	}
	if totalLines > 0 {
		overallPct = (float64(totalCovered) / float64(totalLines)) * 100.0
	}
	return
}

// CalculateCoverage calculates the coverage percentage for a file
func (fc *FileCoverage) CalculateCoverage() {
	if fc.TotalLines > 0 {
		fc.CoveragePct = (float64(fc.CoveredLines) / float64(fc.TotalLines)) * 100.0
	}
}
