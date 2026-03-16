package cmd

import (
	"fmt"
	"os"

	"github.com/FortuneW/gollmperf/internal/utils"
	"github.com/spf13/cobra"
)

var testRandomCmd = &cobra.Command{
	Use:   "test-random",
	Short: "Test random prompt generation for vLLM",
	Long: `Test the random prompt generation functionality.
This command tests the GetRandomPromptByTokenCount function to ensure
the generated prompts match the target token count within tolerance.`,
	Run: func(cmd *cobra.Command, args []string) {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		targetTokens, _ := cmd.Flags().GetInt("tokens")
		iterations, _ := cmd.Flags().GetInt("iterations")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if endpoint == "" {
			endpoint = os.Getenv("LLM_API_ENDPOINT")
			if endpoint == "" {
				mlog.Error("Endpoint must be specified via --endpoint flag or LLM_API_ENDPOINT environment variable")
				os.Exit(1)
			}
		}

		mlog.Infof("Testing random prompt generation")
		mlog.Infof("Endpoint: %s", endpoint)
		mlog.Infof("Target tokens: %d", targetTokens)
		mlog.Infof("Iterations: %d", iterations)
		fmt.Println()

		successCount := 0
		var totalDiff int

		for i := 0; i < iterations; i++ {
			mlog.Infof("=== Iteration %d/%d ===", i+1, iterations)
			prompt, err := utils.GetRandomPromptByTokenCount(endpoint, targetTokens)
			if err != nil {
				mlog.Errorf("Failed to generate prompt: %v", err)
				continue
			}

			// Get actual token count
			tokenizeURL, _ := utils.BuildTokenizeURL(endpoint)
			actualTokens, err := utils.CallTokenizeAPI(tokenizeURL, prompt)
			if err != nil {
				mlog.Errorf("Failed to count tokens: %v", err)
				continue
			}

			diff := actualTokens - targetTokens
			if diff < 0 {
				diff = -diff
			}
			totalDiff += diff

			tolerance := max(10, targetTokens/50)
			status := "PASS"
			if diff > tolerance {
				status = "WARN"
			} else {
				successCount++
			}

			mlog.Infof("Result: %s | Target: %d | Actual: %d | Diff: %d | Tolerance: %d",
				status, targetTokens, actualTokens, diff, tolerance)

			if verbose {
				mlog.Infof("Prompt length: %d chars", len(prompt))
				mlog.Infof("Prompt preview: %s...", truncateString(prompt, 200))
				fmt.Println()
			}
		}

		fmt.Println()
		mlog.Infof("=== Summary ===")
		mlog.Infof("Total iterations: %d", iterations)
		mlog.Infof("Success count (within tolerance): %d", successCount)
		mlog.Infof("Success rate: %.1f%%", float64(successCount)/float64(iterations)*100)
		mlog.Infof("Average diff: %.1f tokens", float64(totalDiff)/float64(iterations))
	},
}

func init() {
	rootCmd.AddCommand(testRandomCmd)
	testRandomCmd.Flags().StringP("endpoint", "e", "", "vLLM endpoint URL (e.g., http://localhost:63535/v1/chat/completions)")
	testRandomCmd.Flags().IntP("tokens", "t", 1000, "Target token count")
	testRandomCmd.Flags().IntP("iterations", "i", 3, "Number of test iterations")
	testRandomCmd.Flags().BoolP("verbose", "v", false, "Show verbose output including prompt preview")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
