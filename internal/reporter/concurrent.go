package reporter

import (
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
