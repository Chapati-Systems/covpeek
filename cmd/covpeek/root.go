package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "covpeek --file <path> [flags]",
	Short: "Cross-language Coverage Report CLI Parser",
	Long: `covpeek is a CLI tool for parsing and analyzing coverage reports 
from multiple languages including Rust, Go, TypeScript, and JavaScript.

It supports LCOV format (.lcov, .info) and Go coverage format (.out).`,
	Example: `  # Parse a coverage file and display summary
  covpeek --file coverage.lcov

  # Force format detection and output as JSON
  covpeek --file coverage.out --format go --output json

  # Filter files below 80% coverage
  covpeek --file coverage.lcov --below 80

  # Output as CSV
  covpeek --file lcov.info --output csv`,
	SilenceUsage: false,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Explicitly set output streams for help and error messages
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}
