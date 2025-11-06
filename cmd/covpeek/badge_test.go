package main

import (
	"os"
	"strings"
	"testing"
)

func TestGetColorForCoverage(t *testing.T) {
	tests := []struct {
		pct    float64
		expect string
	}{
		{95, "#4c1"},
		{90, "#4c1"},
		{89, "#a4a61d"},
		{80, "#a4a61d"},
		{79, "#dfb317"},
		{70, "#dfb317"},
		{69, "#fe7d37"},
		{60, "#fe7d37"},
		{59, "#e05d44"},
		{0, "#e05d44"},
	}

	for _, test := range tests {
		got := getColorForCoverage(test.pct)
		if got != test.expect {
			t.Errorf("getColorForCoverage(%.1f) = %s, expect %s", test.pct, got, test.expect)
		}
	}
}

func TestGenerateBadgeSVG(t *testing.T) {
	svg := generateBadgeSVG("coverage", "85.0%", "#a4a61d", "flat")
	if !strings.Contains(svg, `fill="#a4a61d"`) {
		t.Error("SVG should contain the color")
	}
	if !strings.Contains(svg, "coverage") {
		t.Error("SVG should contain the label")
	}
	if !strings.Contains(svg, "85.0%") {
		t.Error("SVG should contain the value")
	}
	if !strings.Contains(svg, `rx="3"`) {
		t.Error("Flat should have rx=3")
	}

	// Test flat-square
	svg = generateBadgeSVG("coverage", "85.0%", "#a4a61d", "flat-square")
	if !strings.Contains(svg, `rx="0"`) {
		t.Error("Flat-square should have rx=0")
	}

	// Test plastic
	svg = generateBadgeSVG("coverage", "85.0%", "#a4a61d", "plastic")
	if !strings.Contains(svg, `rx="4"`) {
		t.Error("Plastic should have rx=4")
	}
	if !strings.Contains(svg, `<linearGradient`) {
		t.Error("Plastic should have gradient")
	}
}

func TestBadgeCommand(t *testing.T) {
	// Create a temp file for output
	tmpFile, err := os.CreateTemp("", "badge-*.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Assume coverage.out exists
	if _, err := os.Stat("coverage.out"); os.IsNotExist(err) {
		t.Skip("coverage.out not found, skipping integration test")
	}

	// Test with file
	badgeFile = "coverage.out"
	badgeOutput = tmpFile.Name()
	badgeLabel = "coverage"
	badgeStyle = "flat"

	err = runBadge(nil, []string{})
	if err != nil {
		t.Errorf("runBadge failed: %v", err)
	}

	// Check file exists and has content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Error("SVG file is empty")
	}
	if !strings.Contains(string(content), "<svg") {
		t.Error("File does not contain SVG")
	}
}