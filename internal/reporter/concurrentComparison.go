package reporter

import (
	"fmt"

	"github.com/FortuneW/gollmperf/internal/analyzer"
)

// ConcurrentTestResult holds the results of a single concurrent test
type ConcurrentTestResult struct {
	Concurrency int               `json:"concurrency"`
	Metrics     *analyzer.Metrics `json:"metrics"`
}

// ConcurrentComparison holds multiple concurrent test results for comparison
type ConcurrentComparison struct {
	TestResults []ConcurrentTestResult `json:"test_results"`
}

// GetBestQPS returns the test result with the highest QPS
func (c *ConcurrentComparison) GetBestQPS() *ConcurrentTestResult {
	if len(c.TestResults) == 0 {
		return nil
	}

	best := &c.TestResults[0]
	for i := 1; i < len(c.TestResults); i++ {
		if c.TestResults[i].Metrics.QPS > best.Metrics.QPS {
			best = &c.TestResults[i]
		} else if c.TestResults[i].Metrics.QPS == best.Metrics.QPS && c.TestResults[i].Concurrency > best.Concurrency {
			best = &c.TestResults[i]
		}
	}
	return best
}

// GetBestLatency returns the test result with the lowest average latency
func (c *ConcurrentComparison) GetBestLatency() *ConcurrentTestResult {
	if len(c.TestResults) == 0 {
		return nil
	}

	best := &c.TestResults[0]
	for i := 1; i < len(c.TestResults); i++ {
		if best.Metrics.AverageLatency == 0 {
			continue
		}
		if c.TestResults[i].Metrics.AverageLatency < best.Metrics.AverageLatency {
			best = &c.TestResults[i]
		} else if c.TestResults[i].Metrics.AverageLatency == best.Metrics.AverageLatency && c.TestResults[i].Concurrency > best.Concurrency {
			best = &c.TestResults[i]
		}
	}
	return best
}

// GetBestSuccessRate returns the test result with the highest success rate
func (c *ConcurrentComparison) GetBestSuccessRate() *ConcurrentTestResult {
	if len(c.TestResults) == 0 {
		return nil
	}

	best := &c.TestResults[0]
	for i := 1; i < len(c.TestResults); i++ {
		if c.TestResults[i].Metrics.SuccessRate > best.Metrics.SuccessRate {
			best = &c.TestResults[i]
		} else if c.TestResults[i].Metrics.SuccessRate == best.Metrics.SuccessRate && c.TestResults[i].Concurrency > best.Concurrency {
			best = &c.TestResults[i]
		}
	}
	return best
}

// GetBestFirstTokenLatency returns the test result with the lowest average first token latency
func (c *ConcurrentComparison) GetBestFirstTokenLatency() *ConcurrentTestResult {
	if len(c.TestResults) == 0 {
		return nil
	}

	best := &c.TestResults[0]
	for i := 1; i < len(c.TestResults); i++ {
		if c.TestResults[i].Metrics.AverageFirstTokenLatency == 0 {
			continue
		}
		if c.TestResults[i].Metrics.AverageFirstTokenLatency < best.Metrics.AverageFirstTokenLatency {
			best = &c.TestResults[i]
		} else if c.TestResults[i].Metrics.AverageFirstTokenLatency == best.Metrics.AverageFirstTokenLatency && c.TestResults[i].Concurrency > best.Concurrency {
			best = &c.TestResults[i]
		}
	}
	return best
}

// GetBestTokensThroughput returns the test result with the highest throughput
func (c *ConcurrentComparison) GetBestTokensThroughput() *ConcurrentTestResult {
	if len(c.TestResults) == 0 {
		return nil
	}

	best := &c.TestResults[0]
	for i := 1; i < len(c.TestResults); i++ {
		if c.TestResults[i].Metrics.TokensPerSecond > best.Metrics.TokensPerSecond {
			best = &c.TestResults[i]
		} else if c.TestResults[i].Metrics.TokensPerSecond == best.Metrics.TokensPerSecond && c.TestResults[i].Concurrency > best.Concurrency {
			best = &c.TestResults[i]
		}
	}
	return best
}

// GetQPSBottleneck returns the concurrency level at which QPS bottleneck occurs
func (c *ConcurrentComparison) GetQPSBottleneck() *BottleneckResult {
	return c.DetectQPSBottleneck()
}

// DetectLatencyBottleneck is a convenience function that uses default latency-based detection
func (c *ConcurrentComparison) DetectLatencyBottleneck() *BottleneckResult {
	detector := NewLatencyBasedDetector(1.0) // 1.0 threshold for ratio-based detection
	return detector.DetectBottleneck(c.TestResults)
}

// GetLatencyBottleneck returns the concurrency level at which latency bottleneck occurs
func (c *ConcurrentComparison) GetLatencyBottleneck() *BottleneckResult {
	return c.DetectLatencyBottleneck()
}

// DetectFirstTokenLatencyBottleneck is a convenience function that uses default latency-based detection for first token latency
func (c *ConcurrentComparison) DetectFirstTokenLatencyBottleneck() *BottleneckResult {
	detector := NewFirstTokenLatencyBasedDetector(1.0) // 1.0 threshold for ratio-based detection
	return detector.DetectBottleneck(c.TestResults)
}

// GetFirstTokenLatencyBottleneck returns the concurrency level at which first token latency bottleneck occurs
func (c *ConcurrentComparison) GetFirstTokenLatencyBottleneck() *BottleneckResult {
	return c.DetectFirstTokenLatencyBottleneck()
}

// GetRecommendedConcurrency returns the recommended concurrency level based on comprehensive analysis of QPS, Tokens/sec and E2E Latency bottlenecks.
// The function analyzes multiple performance metrics to determine the optimal concurrency level that provides the best balance between
// throughput, latency, and system resource utilization. It considers:
// 1. QPS bottleneck detection - identifies when adding more concurrency no longer improves throughput
// 2. Latency bottleneck detection - identifies when concurrency levels cause unacceptable latency increases
// 3. Best performance metrics - identifies concurrency levels that maximize QPS, token throughput, and minimize latency
//
// The recommendation logic follows these principles:
// - If bottlenecks are detected, recommend staying below the bottleneck level
// - If no bottlenecks are detected, recommend the concurrency level that maximizes overall performance
// - Provide detailed reasoning for the recommendation to help users understand the trade-offs
//
// Returns a BottleneckResult with the recommended concurrency level and detailed reasoning.
func (c *ConcurrentComparison) GetRecommendedConcurrency() *BottleneckResult {
	qpsBottleneck := c.GetQPSBottleneck()
	latencyBottleneck := c.GetLatencyBottleneck()
	bestQPS := c.GetBestQPS()
	bestTokensThroughput := c.GetBestTokensThroughput()
	bestLatency := c.GetBestLatency()

	// Handle edge cases
	if len(c.TestResults) == 0 {
		return &BottleneckResult{
			IsBottleneck:  false,
			AlgorithmUsed: "RecommendedConcurrency",
		}
	}

	if len(c.TestResults) == 1 {
		return &BottleneckResult{
			Concurrency:      c.TestResults[0].Concurrency,
			QPS:              float64(c.TestResults[0].Metrics.QPS),
			TokensPerSec:     float64(c.TestResults[0].Metrics.TokensPerSecond),
			AverageLatencyMS: c.TestResults[0].Metrics.AverageLatency.Milliseconds(),
			IsBottleneck:     false,
			AlgorithmUsed:    "RecommendedConcurrency",
			Reason:           "Only one concurrency level tested",
		}
	}

	// If any of the key metrics are nil, return a default result
	if qpsBottleneck == nil || latencyBottleneck == nil || bestQPS == nil || bestTokensThroughput == nil || bestLatency == nil {
		// Return the best QPS result as fallback
		if bestQPS != nil {
			return &BottleneckResult{
				Concurrency:      bestQPS.Concurrency,
				QPS:              float64(bestQPS.Metrics.QPS),
				TokensPerSec:     float64(bestQPS.Metrics.TokensPerSecond),
				AverageLatencyMS: bestQPS.Metrics.AverageLatency.Milliseconds(),
				IsBottleneck:     false,
				AlgorithmUsed:    "RecommendedConcurrency",
				Reason:           "Fallback to best QPS result",
			}
		}
		// If still no result, return the first test result
		return &BottleneckResult{
			Concurrency:      c.TestResults[0].Concurrency,
			QPS:              float64(c.TestResults[0].Metrics.QPS),
			TokensPerSec:     float64(c.TestResults[0].Metrics.TokensPerSecond),
			AverageLatencyMS: c.TestResults[0].Metrics.AverageLatency.Milliseconds(),
			IsBottleneck:     false,
			AlgorithmUsed:    "RecommendedConcurrency",
			Reason:           "Fallback to first test result",
		}
	}

	// Determine the recommended concurrency based on multiple factors
	recommended := &BottleneckResult{
		Concurrency:      bestQPS.Concurrency,
		QPS:              float64(bestQPS.Metrics.QPS),
		TokensPerSec:     float64(bestQPS.Metrics.TokensPerSecond),
		AverageLatencyMS: bestQPS.Metrics.AverageLatency.Milliseconds(),
		IsBottleneck:     false,
		AlgorithmUsed:    "RecommendedConcurrency",
	}

	// Set the reason based on the analysis
	if qpsBottleneck.IsBottleneck && latencyBottleneck.IsBottleneck {
		// Both QPS and latency show bottleneck
		if qpsBottleneck.Concurrency <= latencyBottleneck.Concurrency {
			// QPS bottleneck occurs at lower concurrency than latency bottleneck
			recommended.Concurrency = qpsBottleneck.Concurrency
			recommended.QPS = qpsBottleneck.QPS
			recommended.TokensPerSec = qpsBottleneck.TokensPerSec
			recommended.AverageLatencyMS = qpsBottleneck.AverageLatencyMS
			recommended.Reason = fmt.Sprintf("QPS bottleneck detected at concurrency %d. Recommend staying below this level for optimal performance.", qpsBottleneck.Concurrency)
		} else {
			// Latency bottleneck occurs at lower concurrency than QPS bottleneck
			recommended.Concurrency = latencyBottleneck.Concurrency
			recommended.QPS = latencyBottleneck.QPS
			recommended.TokensPerSec = latencyBottleneck.TokensPerSec
			recommended.AverageLatencyMS = latencyBottleneck.AverageLatencyMS
			recommended.Reason = fmt.Sprintf("Latency bottleneck detected at concurrency %d. Recommend staying below this level to maintain low latency.", latencyBottleneck.Concurrency)
		}
	} else if qpsBottleneck.IsBottleneck {
		// Only QPS bottleneck detected
		recommended.Concurrency = qpsBottleneck.Concurrency
		recommended.QPS = qpsBottleneck.QPS
		recommended.TokensPerSec = qpsBottleneck.TokensPerSec
		recommended.AverageLatencyMS = qpsBottleneck.AverageLatencyMS

		// Check if the total requests are less than the concurrency level
		// This might indicate that the bottleneck is not genuine
		var totalRequests int
		for _, result := range c.TestResults {
			if result.Concurrency == qpsBottleneck.Concurrency {
				totalRequests = result.Metrics.TotalRequests
				break
			}
		}
		if totalRequests > 0 && totalRequests <= qpsBottleneck.Concurrency*2 {
			recommended.Reason = fmt.Sprintf("QPS bottleneck detected at concurrency %d, but only %d requests were processed. This may indicate the bottleneck is not genuine - consider running longer tests to confirm.", qpsBottleneck.Concurrency, totalRequests)
		} else {
			recommended.Reason = fmt.Sprintf("QPS bottleneck detected at concurrency %d. Recommend staying below this level for optimal throughput.", qpsBottleneck.Concurrency)
		}
	} else if latencyBottleneck.IsBottleneck {
		// Only latency bottleneck detected
		recommended.Concurrency = latencyBottleneck.Concurrency
		recommended.QPS = latencyBottleneck.QPS
		recommended.TokensPerSec = latencyBottleneck.TokensPerSec
		recommended.AverageLatencyMS = latencyBottleneck.AverageLatencyMS
		recommended.Reason = fmt.Sprintf("Latency bottleneck detected at concurrency %d. Recommend staying below this level to maintain low latency.", latencyBottleneck.Concurrency)
	} else {
		// No clear bottleneck detected, recommend the best overall concurrency
		// Prefer the concurrency that maximizes both QPS and Tokens/sec while maintaining reasonable latency
		if bestTokensThroughput.Concurrency == bestQPS.Concurrency {
			recommended.Concurrency = bestQPS.Concurrency
			recommended.QPS = float64(bestQPS.Metrics.QPS)
			recommended.TokensPerSec = float64(bestQPS.Metrics.TokensPerSecond)
			recommended.AverageLatencyMS = bestQPS.Metrics.AverageLatency.Milliseconds()
			recommended.Reason = fmt.Sprintf("Optimal concurrency %d maximizes both QPS and token throughput.", bestQPS.Concurrency)
		} else if bestLatency.Concurrency <= bestQPS.Concurrency && bestLatency.Concurrency <= bestTokensThroughput.Concurrency {
			// If best latency is at lower concurrency, recommend a balance
			recommended.Concurrency = bestQPS.Concurrency
			recommended.QPS = float64(bestQPS.Metrics.QPS)
			recommended.TokensPerSec = float64(bestQPS.Metrics.TokensPerSecond)
			recommended.AverageLatencyMS = bestQPS.Metrics.AverageLatency.Milliseconds()
			recommended.Reason = fmt.Sprintf("Recommended concurrency %d for balanced performance between QPS and latency.", bestQPS.Concurrency)
		} else {
			// Default to best QPS
			recommended.Concurrency = bestQPS.Concurrency
			recommended.QPS = float64(bestQPS.Metrics.QPS)
			recommended.TokensPerSec = float64(bestQPS.Metrics.TokensPerSecond)
			recommended.AverageLatencyMS = bestQPS.Metrics.AverageLatency.Milliseconds()
			recommended.Reason = fmt.Sprintf("Recommended concurrency %d for maximum QPS.", bestQPS.Concurrency)
		}
	}

	return recommended
}
