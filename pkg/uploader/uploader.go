package uploader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Uploader interface for uploading coverage reports to different platforms
type Uploader interface {
	Upload(filePath string) error
}

// NewSonarQubeUploader creates a new SonarQube uploader
func NewSonarQubeUploader(url, token, projectKey string, verbose, quiet bool) (Uploader, error) {
	return &sonarQubeUploader{
		url:        url,
		token:      token,
		projectKey: projectKey,
		verbose:    verbose,
		quiet:      quiet,
	}, nil
}

// NewCodecovUploader creates a new Codecov uploader
func NewCodecovUploader(token string, verbose, quiet bool) (Uploader, error) {
	return &codecovUploader{
		token:   token,
		verbose: verbose,
		quiet:   quiet,
	}, nil
}

// sonarQubeUploader implements Uploader for SonarQube
type sonarQubeUploader struct {
	url        string
	token      string
	projectKey string
	verbose    bool
	quiet      bool
}

func (s *sonarQubeUploader) Upload(filePath string) error {
	if !s.quiet {
		fmt.Printf("Uploading to SonarQube at %s for project %s\n", s.url, s.projectKey)
	}

	// Check if sonar-scanner is available
	if _, err := exec.LookPath("sonar-scanner"); err != nil {
		return fmt.Errorf("sonar-scanner not found in PATH. Please install SonarQube Scanner")
	}

	// Create temporary sonar-project.properties
	props := fmt.Sprintf(`sonar.projectKey=%s
sonar.host.url=%s
sonar.login=%s
sonar.coverage.jacoco.xmlReportPaths=%s
`, s.projectKey, s.url, s.token, filePath)

	tempDir := os.TempDir()
	propsFile := filepath.Join(tempDir, "sonar-project.properties")
	if err := os.WriteFile(propsFile, []byte(props), 0644); err != nil {
		return fmt.Errorf("failed to create sonar-project.properties: %w", err)
	}

	// Run sonar-scanner
	cmd := exec.Command("sonar-scanner", "-Dproject.settings="+propsFile)
	if s.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SonarQube scan failed: %w", err)
	}

	return nil
}

// codecovUploader implements Uploader for Codecov
type codecovUploader struct {
	token   string
	verbose bool
	quiet   bool
}

func (c *codecovUploader) Upload(filePath string) error {
	if !c.quiet {
		fmt.Println("Uploading to Codecov")
	}

	// Download the Codecov CLI
	cliPath, err := c.downloadCodecovCLI()
	if err != nil {
		return fmt.Errorf("failed to download Codecov CLI: %w", err)
	}

	// Run the CLI
	args := []string{"upload-process", "-t", c.token, "-f", filePath}
	if c.verbose {
		args = append(args, "--verbose")
	}

	cmd := exec.Command(cliPath, args...)
	if c.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codecov upload failed: %w", err)
	}

	return nil
}

func (c *codecovUploader) downloadCodecovCLI() (string, error) {
	// Determine OS and arch
	osName := runtime.GOOS
	arch := runtime.GOARCH

	var distro string
	switch arch {
	case "amd64":
		switch osName {
		case "linux":
			distro = "linux"
		case "darwin":
			distro = "macos"
		case "windows":
			distro = "windows"
		default:
			return "", fmt.Errorf("unsupported OS: %s", osName)
		}
	case "arm64":
		switch osName {
		case "linux":
			distro = "linux-arm64"
		case "darwin":
			distro = "macos"
		default:
			return "", fmt.Errorf("unsupported OS: %s", osName)
		}
	default:
		return "", fmt.Errorf("unsupported arch: %s", arch)
	}

	url := fmt.Sprintf("https://cli.codecov.io/latest/%s/codecov", distro)

	// Create temp file
	tempDir := os.TempDir()
	cliPath := filepath.Join(tempDir, "codecov-cli")
	if osName == "windows" {
		cliPath += ".exe"
	}

	// Download
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download CLI: %s", resp.Status)
	}

	out, err := os.Create(cliPath)
	if err != nil {
		return "", err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	// Make executable
	if osName != "windows" {
		if err := os.Chmod(cliPath, 0755); err != nil {
			return "", err
		}
	}

	return cliPath, nil
}
