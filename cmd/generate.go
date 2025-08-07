package cmd

import (
	"github.com/FortuneW/gollmperf/internal/config"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [output_file] ",
	Short: "Generate default configuration file",
	Long:  `Generate a default configuration file with reasonable defaults.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputPath := args[0]
		if err := config.GenerateDefaultConfig(outputPath); err != nil {
			mlog.Errorf("Failed to generate default config: %v", err)
			return
		}
		mlog.Infof("Default configuration file generated successfully: %s", outputPath)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
