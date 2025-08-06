package cmd

import (
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare performance between models",
	Long:  `Compare performance between different models or configurations`,
	Run: func(cmd *cobra.Command, args []string) {
		mlog.Info("Running comparison test...")
		// Implementation will be added later
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)
	compareCmd.Flags().StringSliceP("configs", "c", nil, "Configuration files to compare")
}
