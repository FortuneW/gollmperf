package reporter

import (
	"fmt"
	"os"
)

// GenerateCSVReport generates a CSV report
func (r *Reporter) GenerateCSVReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write CSV header
	header := "concurrency,total_requests,successful_requests,failed_requests,success_rate,qps,tokens_per_second," +
		"average_latency,latency_p50,latency_p90,latency_p99," +
		"average_request_tokens,average_response_tokens," +
		"average_first_token_latency,first_token_latency_p50,first_token_latency_p90,first_token_latency_p99\n"

	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, result := range r.concurrentComparison.TestResults {
		// Write data row
		row := fmt.Sprintf("%d,%d,%d,%d,%.2f,%.2f,%.2f,%d,%d,%d,%d,%.2f,%.2f,%d,%d,%d,%d\n",
			result.Concurrency,
			result.Metrics.TotalRequests,
			result.Metrics.SuccessfulRequests,
			result.Metrics.FailedRequests,
			result.Metrics.SuccessRate,
			result.Metrics.QPS,
			result.Metrics.TokensPerSecond,
			result.Metrics.AverageLatency.Milliseconds(),
			result.Metrics.LatencyP50.Milliseconds(),
			result.Metrics.LatencyP90.Milliseconds(),
			result.Metrics.LatencyP99.Milliseconds(),
			result.Metrics.AverageRequestTokens,
			result.Metrics.AverageResponseTokens,
			result.Metrics.AverageFirstTokenLatency.Milliseconds(),
			result.Metrics.FirstTokenLatencyP50.Milliseconds(),
			result.Metrics.FirstTokenLatencyP90.Milliseconds(),
			result.Metrics.FirstTokenLatencyP99.Milliseconds(),
		)

		if _, err := file.WriteString(row); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}
	return nil
}
