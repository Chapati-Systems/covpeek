package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunUploadValidation(t *testing.T) {
	// Test invalid platform
	uploadTo = "invalid"
	uploadFile = "../../coverage.out"
	err := runUpload(nil, []string{})
	if err == nil || err.Error() != "invalid platform 'invalid': must be 'sonarqube' or 'codecov'" {
		t.Errorf("Expected error for invalid platform, got %v", err)
	}

	// Test missing project-key for sonarqube
	uploadTo = "sonarqube"
	projectKey = ""
	sonarToken = "test"
	err = runUpload(nil, []string{})
	if err == nil || err.Error() != "--project-key is required for SonarQube" {
		t.Errorf("Expected error for missing project-key, got %v", err)
	}

	// Test missing token for codecov
	uploadTo = "codecov"
	repoToken = ""
	_ = os.Unsetenv("CODECOV_TOKEN")
	err = runUpload(nil, []string{})
	if err == nil || err.Error() != "repository token required. Use --repo-token or set CODECOV_TOKEN env var" {
		t.Errorf("Expected error for missing token, got %v", err)
	}
}

func TestRunUploadAutoDetectMultiple(t *testing.T) {
	// Create multiple temp files
	files := []string{"coverage.out", "lcov.info"}
	created := []string{}

	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			err := os.WriteFile(f, []byte("dummy"), 0644)
			if err != nil {
				t.Fatalf("Failed to create %s: %v", f, err)
			}
			created = append(created, f)
		}
	}

	defer func() {
		for _, f := range created {
			_ = os.Remove(f)
		}
	}()

	uploadTo = "codecov"
	repoToken = "test"
	uploadFile = ""
	err := runUpload(nil, []string{})
	if err == nil || !strings.Contains(err.Error(), "multiple coverage files detected") {
		t.Errorf("Expected multiple files error, got %v", err)
	}
}

func TestRunUploadInvalidFile(t *testing.T) {
	uploadTo = "codecov"
	repoToken = "test"
	uploadFile = "nonexistent"
	err := runUpload(nil, []string{})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error for invalid file, got %v", err)
	}
}
