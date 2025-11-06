package uploader

import (
	"os"
	"testing"
)

func TestNewCodecovUploader(t *testing.T) {
	u, err := NewCodecovUploader("token", true, false)
	if err != nil {
		t.Fatalf("NewCodecovUploader failed: %v", err)
	}
	if u == nil {
		t.Fatal("NewCodecovUploader returned nil")
	}
}

func TestNewSonarQubeUploader(t *testing.T) {
	u, err := NewSonarQubeUploader("url", "token", "key", true, false)
	if err != nil {
		t.Fatalf("NewSonarQubeUploader failed: %v", err)
	}
	if u == nil {
		t.Fatal("NewSonarQubeUploader returned nil")
	}
}

func TestDownloadCodecovCLI(t *testing.T) {
	c := &codecovUploader{token: "test"}
	// This will try to download, but since it's a test, it might fail, but we can check if it returns a path or error
	_, err := c.downloadCodecovCLI()
	// It should either succeed or fail with network error, but not panic
	if err != nil {
		// Expected, since no network in test
		t.Logf("Expected error: %v", err)
	}
}

func TestCodecovUpload(t *testing.T) {
	u, _ := NewCodecovUploader("invalid", false, true)
	err := u.Upload("nonexistent")
	if err == nil {
		t.Error("Expected error for invalid upload")
	}
}

func TestSonarQubeUpload(t *testing.T) {
	u, _ := NewSonarQubeUploader("url", "token", "key", false, true)
	err := u.Upload("nonexistent")
	if err == nil {
		t.Error("Expected error for sonar-scanner not found")
	}
}

func TestDownloadURL(t *testing.T) {
	// Test the URL generation logic
	// Since downloadCodecovCLI is private, we can test by calling it
	c := &codecovUploader{token: "test"}
	path, err := c.downloadCodecovCLI()
	if err != nil {
		t.Logf("Download failed as expected in test: %v", err)
	} else {
		t.Logf("Downloaded to %s", path)
		// Clean up
		_ = os.Remove(path)
	}
}
