package analyzer

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/FortuneW/gollmperf/internal/collector"
)

// Duration is a wrapper around time.Duration that marshals to milliseconds in JSON
type Duration time.Duration

// MarshalJSON implements json.Marshaler interface
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).Milliseconds())
}

// Seconds returns the duration as a floating point number of seconds.
func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}

// Milliseconds returns the duration as an integer millisecond count.
func (d Duration) Milliseconds() int64 {
	return time.Duration(d).Milliseconds()
}

func (d Duration) String() string {
	return fmt.Sprintf("%v", time.Duration(d))
}

// Float64 is a wrapper around float64 that marshals to 3 decimal places in JSON
type Float64 float64

// MarshalJSON implements json.Marshaler interface
func (f Float64) MarshalJSON() ([]byte, error) {
	return json.Marshal(math.Round(float64(f)*1000) / 1000)
}

// Metrics represents the calculated performance metrics
type Metrics struct {
	// Basic metrics
	TotalRequests      int     `json:"total_requests"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	SuccessRate        Float64 `json:"success_rate"`

	// Timing metrics
	TotalDuration  Duration `json:"total_duration"`
	AverageLatency Duration `json:"average_latency"`
	LatencyP50     Duration `json:"latency_p50"`
	LatencyP90     Duration `json:"latency_p90"`
	LatencyP99     Duration `json:"latency_p99"`

	// Throughput metrics
	QPS             Float64 `json:"qps"`
	TotalTokens     int     `json:"total_tokens"`
	TokensPerSecond Float64 `json:"tokens_per_second"`

	// Token metrics
	AverageRequestTokens  Float64 `json:"average_request_tokens"`
	AverageResponseTokens Float64 `json:"average_response_tokens"`

	// Streaming metrics (if applicable)
	AverageFirstTokenLatency Duration `json:"average_first_token_latency,omitempty"`
	FirstTokenLatencyP50     Duration `json:"first_token_latency_p50,omitempty"`
	FirstTokenLatencyP90     Duration `json:"first_token_latency_p90,omitempty"`
	FirstTokenLatencyP99     Duration `json:"first_token_latency_p99,omitempty"`

	// Error analysis
	ErrorCounts     map[string]int `json:"error_counts"`
	ErrorTypeCounts map[string]int `json:"error_type_counts,omitempty"`
}

// Analyzer analyzes test results and calculates metrics
type Analyzer struct {
	collector *collector.Collector
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(col *collector.Collector) *Analyzer {
	return &Analyzer{
		collector: col,
	}
}

// Analyze performs analysis on the collected results
func (a *Analyzer) Analyze() *Metrics {
	results := a.collector.GetAllResults()
	successfulResults := a.collector.GetSuccessfulResults()

	metrics := &Metrics{
		ErrorCounts:     make(map[string]int),
		ErrorTypeCounts: make(map[string]int),
	}

	// Basic metrics
	metrics.TotalRequests = len(results)
	metrics.SuccessfulRequests = len(successfulResults)
	metrics.FailedRequests = a.collector.GetFailureCount()

	if metrics.TotalRequests > 0 {
		metrics.SuccessRate = Float64(metrics.SuccessfulRequests) / Float64(metrics.TotalRequests) * 100
	}

	// Duration
	metrics.TotalDuration = Duration(a.collector.GetTestDuration())
	if metrics.TotalDuration < 0 {
		metrics.TotalDuration = 0
	}

	// Calculate QPS
	if metrics.TotalDuration > 0 {
		metrics.QPS = Float64(metrics.SuccessfulRequests) / Float64(metrics.TotalDuration.Seconds())
	}

	// Only calculate detailed metrics if we have successful results
	if len(successfulResults) > 0 {
		// Latency calculations
		latencies := make([]time.Duration, len(successfulResults))
		firstTokenLatencies := make([]time.Duration, 0)
		totalLatency := time.Duration(0)
		totalRequestTokens := 0
		totalResponseTokens := 0

		for i, result := range successfulResults {
			latencies[i] = result.Latency
			totalLatency += result.Latency
			totalRequestTokens += result.RequestTokens
			totalResponseTokens += result.ResponseTokens

			// Collect first token latencies if available
			if result.FirstTokenLatency > 0 {
				firstTokenLatencies = append(firstTokenLatencies, result.FirstTokenLatency)
			}
		}

		// Average latency
		metrics.AverageLatency = Duration(totalLatency / time.Duration(len(successfulResults)))

		// Sort latencies for percentile calculations
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		// Latency percentiles
		metrics.LatencyP50 = Duration(latencies[int(float64(len(latencies))*0.5)])
		metrics.LatencyP90 = Duration(latencies[int(float64(len(latencies))*0.9)])
		metrics.LatencyP99 = Duration(latencies[int(float64(len(latencies))*0.99)])

		// Token metrics
		metrics.TotalTokens = totalRequestTokens + totalResponseTokens
		metrics.AverageRequestTokens = Float64(totalRequestTokens) / Float64(len(successfulResults))
		metrics.AverageResponseTokens = Float64(totalResponseTokens) / Float64(len(successfulResults))

		// Tokens per second
		if metrics.TotalDuration > 0 {
			metrics.TokensPerSecond = Float64(metrics.TotalTokens) / Float64(metrics.TotalDuration.Seconds())
		}

		// First token latency metrics (if available)
		if len(firstTokenLatencies) > 0 {
			totalFirstTokenLatency := time.Duration(0)
			for _, lat := range firstTokenLatencies {
				totalFirstTokenLatency += lat
			}
			metrics.AverageFirstTokenLatency = Duration(totalFirstTokenLatency / time.Duration(len(firstTokenLatencies)))

			// Sort first token latencies for percentile calculations
			sort.Slice(firstTokenLatencies, func(i, j int) bool {
				return firstTokenLatencies[i] < firstTokenLatencies[j]
			})

			// First token latency percentiles
			metrics.FirstTokenLatencyP50 = Duration(firstTokenLatencies[int(float64(len(firstTokenLatencies))*0.5)])
			metrics.FirstTokenLatencyP90 = Duration(firstTokenLatencies[int(float64(len(firstTokenLatencies))*0.9)])
			metrics.FirstTokenLatencyP99 = Duration(firstTokenLatencies[int(float64(len(firstTokenLatencies))*0.99)])
		}
	}

	// Error analysis
	failedResults := a.collector.GetFailedResults()
	for _, result := range failedResults {
		metrics.ErrorCounts[result.ErrorType]++
	}

	// Error type analysis
	for _, result := range failedResults {
		if result.ErrorType != "" {
			metrics.ErrorTypeCounts[result.ErrorType]++
		} else {
			// Default to "unknown" for results without error type
			metrics.ErrorTypeCounts["unknown"]++
		}
	}

	return metrics
}
