package cmd

import (
	"context"

	"github.com/feildrix/solami/internal/app"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Solami terminal wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return app.Run(context.Background())
	}
}
