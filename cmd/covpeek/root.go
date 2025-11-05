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
	Example: `  # Parse a coverage file and display table
  covpeek --file coverage.lcov

  # Launch interactive TUI
  covpeek --file coverage.lcov --tui

  # Output as JSON
  covpeek --file coverage.out --output json

  # Filter files below 80% coverage
  covpeek --file coverage.lcov --below 80

  # Force format detection
  covpeek --file coverage.txt --force-format lcov`,
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
