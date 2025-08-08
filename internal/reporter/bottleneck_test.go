package reporter

import (
	"testing"

	"github.com/FortuneW/gollmperf/internal/analyzer"
)

func TestGradientBasedDetector_DetectBottleneck(t *testing.T) {
	// Create test results with clear bottleneck pattern
	results := []ConcurrentTestResult{
		{
			Concurrency: 1,
			Metrics: &analyzer.Metrics{
				QPS:                10.0,
				SuccessfulRequests: 100,
			},
		},
		{
			Concurrency: 2,
			Metrics: &analyzer.Metrics{
				QPS:                20.0,
				SuccessfulRequests: 200,
			},
		},
		{
			Concurrency: 3,
			Metrics: &analyzer.Metrics{
				QPS:                25.0,
				SuccessfulRequests: 250,
			},
		},
		{
			Concurrency: 4,
			Metrics: &analyzer.Metrics{
				QPS:                27.0,
				SuccessfulRequests: 270,
			},
		},
		{
			Concurrency: 5,
			Metrics: &analyzer.Metrics{
				QPS:                27.5,
				SuccessfulRequests: 275,
			},
		},
	}

	// Test with 0.1 threshold (10%)
	detector := NewGradientBasedDetector(0.1)
	_ = detector.DetectBottleneck(results)

	// Test with very high threshold (should detect bottleneck)
	detectorHigh := NewGradientBasedDetector(10.0)
	resultHigh := detectorHigh.DetectBottleneck(results)
	
	// With such a high threshold, a bottleneck should be detected at concurrency 2
	// because the gradient from 2->3 is 5.0, which is less than 10.0
	if !resultHigh.IsBottleneck {
		t.Error("Expected bottleneck with high threshold")
	}
	if resultHigh.Concurrency != 2 {
		t.Errorf("Expected bottleneck at concurrency 2, got %d", resultHigh.Concurrency)
	}

	// Test with empty results
	emptyResults := []ConcurrentTestResult{}
	resultEmpty := detector.DetectBottleneck(emptyResults)
	if resultEmpty.IsBottleneck {
		t.Error("Expected no bottleneck with empty results")
	}

	// Test with single result
	singleResult := []ConcurrentTestResult{
		{
			Concurrency: 1,
			Metrics: &analyzer.Metrics{
				QPS:                10.0,
				SuccessfulRequests: 100,
			},
		},
	}
	resultSingle := detector.DetectBottleneck(singleResult)
	if resultSingle.IsBottleneck {
		t.Error("Expected no bottleneck with single result")
	}
}

func TestStatisticalBasedDetector_DetectBottleneck(t *testing.T) {
	// Create test results with clear bottleneck pattern
	results := []ConcurrentTestResult{
		{
			Concurrency: 1,
			Metrics: &analyzer.Metrics{
				QPS:                10.0,
				SuccessfulRequests: 100,
			},
		},
		{
			Concurrency: 2,
			Metrics: &analyzer.Metrics{
				QPS:                20.0,
				SuccessfulRequests: 200,
			},
		},
		{
			Concurrency: 3,
			Metrics: &analyzer.Metrics{
				QPS:                30.0,
				SuccessfulRequests: 300,
			},
		},
		{
			Concurrency: 4,
			Metrics: &analyzer.Metrics{
				QPS:                31.0,
				SuccessfulRequests: 310,
			},
		},
		{
			Concurrency: 5,
			Metrics: &analyzer.Metrics{
				QPS:                31.5,
				SuccessfulRequests: 315,
			},
		},
		{
			Concurrency: 6,
			Metrics: &analyzer.Metrics{
				QPS:                31.8,
				SuccessfulRequests: 318,
			},
		},
	}

	// Test with window size 3 and threshold 0.05
	detector := NewStatisticalBasedDetector(3, 0.05)
	result := detector.DetectBottleneck(results)

	// Should detect bottleneck where coefficient of variation drops below threshold
	if result.AlgorithmUsed != "StatisticalBased" {
		t.Errorf("Expected StatisticalBased algorithm, got %s", result.AlgorithmUsed)
	}

	// Test with insufficient results
	insufficientResults := results[:2] // Only 2 results
	resultInsufficient := detector.DetectBottleneck(insufficientResults)
	if resultInsufficient.IsBottleneck {
		t.Error("Expected no bottleneck with insufficient results")
	}
}

func TestConcurrentComparison_GetQPSBottleneck(t *testing.T) {
	// Create test results
	results := []ConcurrentTestResult{
		{
			Concurrency: 1,
			Metrics: &analyzer.Metrics{
				QPS:                10.0,
				SuccessfulRequests: 100,
			},
		},
		{
			Concurrency: 2,
			Metrics: &analyzer.Metrics{
				QPS:                20.0,
				SuccessfulRequests: 200,
			},
		},
	}

	// Create ConcurrentComparison with test results
	cc := &ConcurrentComparison{
		TestResults: results,
	}

	// Test the convenience method
	result := cc.GetQPSBottleneck()
	if result == nil {
		t.Error("Expected non-nil result")
	}
	
	if result.AlgorithmUsed != "GradientBased" {
		t.Errorf("Expected GradientBased algorithm, got %s", result.AlgorithmUsed)
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}