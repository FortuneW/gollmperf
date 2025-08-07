package cmd

import (
	"fmt"
	"os"

	"github.com/FortuneW/gollmperf/internal/analyzer"
	"github.com/FortuneW/gollmperf/internal/collector"
	"github.com/FortuneW/gollmperf/internal/engine"
	"github.com/FortuneW/gollmperf/internal/reporter"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run batch or stress test or perf test",
	Long: `Run batch mode to finish all cases;
Run stress mode test to find system stability;
Run perf mode test to find performance limits in different concurrency levels`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize test context
		testCtx := InitializeTest(runFlags)

		// Create reporter
		r := reporter.NewReporter()

		runOnceTest := func(ctx *TestContext, isStress bool) {
			// Run test and get collector
			col, err := runTest(ctx, isStress)
			if err != nil {
				mlog.Errorf("Failed to run test (stress mode: %v): %v", isStress, err)
				os.Exit(1)
			}
			// Analyze results
			resultAnalyzer := analyzer.NewAnalyzer(col)
			// Get metrics
			metrics := resultAnalyzer.Analyze()

			// Generate console report
			r.AddNewMetrics(testCtx.Config.Test.Concurrency, metrics)
			r.GenerateConsoleReport()

			// Generate file report if requested
			if err := r.GenerateFileReport(testCtx.Config.Output.Path, testCtx.Config.Output.Format); err != nil {
				mlog.Errorf("failed to generate file report [%s]: %v", testCtx.Config.Output.Path, err)
			}
		}

		if !runFlags.IsPerf {
			runOnceTest(testCtx, runFlags.IsStress)
		} else {
			// Run perf test
			mlog.Infof("Running perf mode with concurrency group: %v", testCtx.Config.Test.PerfConcurrencyGroup)
			for _, concurrency := range testCtx.Config.Test.PerfConcurrencyGroup {
				testCtx.Config.Test.Concurrency = concurrency
				runOnceTest(testCtx, runFlags.IsStress)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&runFlags.IsStress, "stress", "s", false, "Run stress mode")
	runCmd.Flags().BoolVarP(&runFlags.IsPerf, "perf", "p", false, "Run perf mode")
	runCmd.Flags().StringVarP(&runFlags.ConfigPath, "config", "c", "", "config file (default is ./example.yaml)")
	runCmd.Flags().StringVarP(&runFlags.Provider, "provider", "P", "openai", "LLM provider (openai, qwen, etc.)")
	runCmd.Flags().StringVarP(&runFlags.Model, "model", "m", "", "Model name")
	runCmd.Flags().StringVarP(&runFlags.Dataset, "dataset", "d", "", "Dataset file path")
	runCmd.Flags().StringVarP(&runFlags.ApiKey, "apikey", "k", "", "API key")
	runCmd.Flags().StringVarP(&runFlags.Endpoint, "endpoint", "e", "", "Endpoint")
	runCmd.Flags().StringVarP(&runFlags.ReportFile, "report", "r", "", "Report file path (output report to file)")
	runCmd.Flags().StringVarP(&runFlags.ReportFormat, "format", "f", "", "Report format (json, csv, html) (default as report file extension)")
}

// runTest executes the test based on the test context and mode
func runTest(testCtx *TestContext, isStress bool) (*collector.Collector, error) {
	// Create engine
	testEngine := engine.NewEngine(testCtx.Config, testCtx.Provider)

	// Run Test
	if isStress {
		mlog.Infof("Running stress mode with provider: %s [%s], model: [%s]",
			testCtx.Config.Model.Provider, testCtx.Config.Model.Endpoint, testCtx.Config.Model.Name)
		results, err := testEngine.RunStress(testCtx.Dataset)
		if err != nil {
			return nil, fmt.Errorf("stress test failed: %w", err)
		}
		return collector.NewCollector(results), nil
	} else {
		mlog.Infof("Running batch mode with provider: %s [%s], model: [%s]",
			testCtx.Config.Model.Provider, testCtx.Config.Model.Endpoint, testCtx.Config.Model.Name)
		results, err := testEngine.RunBatch(testCtx.Dataset)
		if err != nil {
			return nil, fmt.Errorf("batch test failed: %w", err)
		}
		return collector.NewCollector(results), nil
	}
}
