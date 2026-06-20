package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "solami",
	Short: "Solami is a TUI-first multi-chain crypto wallet",
}

// Execute runs the Solami CLI.
func Execute() error {
	return rootCmd.Execute()
}
