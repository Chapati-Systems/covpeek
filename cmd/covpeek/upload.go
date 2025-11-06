package main

import (
	"fmt"
	"os"

	"github.com/Chapati-Systems/covpeek/pkg/uploader"
	"github.com/spf13/cobra"
)

var (
	uploadTo   string
	uploadFile string
	projectKey string
	repoToken  string
	sonarURL   string
	sonarToken string
	verbose    bool
	quiet      bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload --to <platform> [flags]",
	Short: "Upload parsed coverage reports to platforms like SonarQube or Codecov",
	Long: `Upload coverage reports directly to popular platforms for integration 
into existing quality and CI dashboards.`,
	Example: `  # Upload to SonarQube
  covpeek upload --to sonarqube --project-key myproj --file coverage/lcov.info --token $SONAR_TOKEN

  # Upload to Codecov
  covpeek upload --to codecov --repo-token $CODECOV_TOKEN`,
	RunE: runUpload,
}

func init() {
	uploadCmd.Flags().StringVar(&uploadTo, "to", "", "Target platform (sonarqube or codecov)")
	uploadCmd.Flags().StringVar(&uploadFile, "file", "", "Path to coverage report file")
	uploadCmd.Flags().StringVar(&projectKey, "project-key", "", "Project key for SonarQube")
	uploadCmd.Flags().StringVar(&repoToken, "repo-token", "", "Repository token for Codecov")
	uploadCmd.Flags().StringVar(&sonarURL, "url", "https://sonarcloud.io", "SonarQube server URL")
	uploadCmd.Flags().StringVar(&sonarToken, "token", "", "Authentication token")
	uploadCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	uploadCmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress output except errors")
	if err := uploadCmd.MarkFlagRequired("to"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	// Validate platform
	if uploadTo != "sonarqube" && uploadTo != "codecov" {
		return fmt.Errorf("invalid platform '%s': must be 'sonarqube' or 'codecov'", uploadTo)
	}

	// Auto-detect file if not specified
	if uploadFile == "" {
		existingFiles := detectExistingCoverageFiles()
		if len(existingFiles) == 0 {
			return fmt.Errorf("no coverage files detected in standard locations. Please specify --file")
		}
		if len(existingFiles) > 1 {
			return fmt.Errorf("multiple coverage files detected: %v. Please specify --file", existingFiles)
		}
		uploadFile = existingFiles[0]
		if !quiet {
			fmt.Fprintf(os.Stderr, "Auto-detected coverage file: %s\n", uploadFile)
		}
	}

	// Validate file exists
	if _, err := os.Stat(uploadFile); os.IsNotExist(err) {
		return fmt.Errorf("coverage file does not exist: %s", uploadFile)
	}

	// Get token from env if not provided
	if uploadTo == "sonarqube" && sonarToken == "" {
		sonarToken = os.Getenv("SONAR_TOKEN")
	}
	if uploadTo == "codecov" && repoToken == "" {
		repoToken = os.Getenv("CODECOV_TOKEN")
	}

	// Validate required auth
	if uploadTo == "sonarqube" {
		if projectKey == "" {
			return fmt.Errorf("--project-key is required for SonarQube")
		}
		if sonarToken == "" {
			return fmt.Errorf("authentication token required. Use --token or set SONAR_TOKEN env var")
		}
	}
	if uploadTo == "codecov" {
		if repoToken == "" {
			return fmt.Errorf("repository token required. Use --repo-token or set CODECOV_TOKEN env var")
		}
	}

	// Create uploader
	var u uploader.Uploader
	var err error
	if uploadTo == "sonarqube" {
		u, err = uploader.NewSonarQubeUploader(sonarURL, sonarToken, projectKey, verbose, quiet)
	} else {
		u, err = uploader.NewCodecovUploader(repoToken, verbose, quiet)
	}
	if err != nil {
		return fmt.Errorf("failed to create uploader: %w", err)
	}

	// Upload
	if err := u.Upload(uploadFile); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	if !quiet {
		fmt.Println("Upload completed successfully")
	}

	return nil
}
