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

func TestRunBadgeInvalidStyle(t *testing.T) {
	badgeFile = "../../testdata/sample.lcov"
	badgeOutput = "test.svg"
	badgeLabel = "coverage"
	badgeStyle = "invalid"

	err := runBadge(nil, []string{})
	if err == nil || !strings.Contains(err.Error(), "must be one of") {
		t.Errorf("Expected error for invalid style, got %v", err)
	}
}

func TestBadgeCommandAutoDetect(t *testing.T) {
	// Create a temp coverage file
	tmpFile := "coverage.out"
	err := os.WriteFile(tmpFile, []byte("mode: set\ngithub.com/example/main.go:10.1,12.1 1 1\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// Create temp output file
	tmpOutput := "test-badge.svg"
	defer func() { _ = os.Remove(tmpOutput) }()

	// Set flags for auto-detect
	badgeFile = ""
	badgeOutput = tmpOutput
	badgeLabel = "coverage"
	badgeStyle = "flat"

	err = runBadge(nil, []string{})
	if err != nil {
		t.Errorf("runBadge auto-detect failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(tmpOutput); os.IsNotExist(err) {
		t.Error("Badge file was not created")
	}
}
