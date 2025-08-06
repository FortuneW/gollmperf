package cmd

import (
	"github.com/FortuneW/qlog"
	"github.com/spf13/cobra"
)

var mlog = qlog.GetRLog("cmd")

var rootCmd = &cobra.Command{
	Use:   "llmperf",
	Short: "Professional LLM performance testing tool",
	Long: `LLMPerf is a professional tool for testing (batch/stress) and benchmarking 
large language models performance with accuracy and precision.`,
}

func Execute() error {
	loglevel, _ := rootCmd.PersistentFlags().GetString("loglevel")
	if loglevel != "" {
		qlog.SetLogLevelStr(loglevel)
	}
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("loglevel", "l", "", "log level")
}
