package reporter

import (
	"math"
)

// BottleneckDetector defines the interface for bottleneck detection algorithms
type BottleneckDetector interface {
	// DetectBottleneck finds the concurrency level at which QPS bottleneck occurs
	DetectBottleneck(results []ConcurrentTestResult) *BottleneckResult
}

// BottleneckResult holds the result of bottleneck detection
type BottleneckResult struct {
	// Concurrency is the concurrency level at which bottleneck occurs
	Concurrency int

	// QPS is the QPS value at bottleneck
	QPS float64

	// IsBottleneck indicates whether a bottleneck was detected
	IsBottleneck bool

	// AlgorithmUsed is the name of the algorithm used for detection
	AlgorithmUsed string
}

// GradientBasedDetector implements bottleneck detection using gradient method
type GradientBasedDetector struct {
	// Threshold is the minimum gradient change to consider as bottleneck
	Threshold float64
}

// NewGradientBasedDetector creates a new gradient-based bottleneck detector
func NewGradientBasedDetector(threshold float64) *GradientBasedDetector {
	return &GradientBasedDetector{
		Threshold: threshold,
	}
}

// DetectBottleneck finds the concurrency level at which QPS bottleneck occurs using gradient method
func (g *GradientBasedDetector) DetectBottleneck(results []ConcurrentTestResult) *BottleneckResult {

	if len(results) == 0 {
		return &BottleneckResult{
			IsBottleneck:  false,
			AlgorithmUsed: "GradientBased",
		}
	} else if len(results) == 1 {
		return &BottleneckResult{
			Concurrency:   results[0].Concurrency,
			QPS:           float64(results[0].Metrics.QPS),
			IsBottleneck:  false,
			AlgorithmUsed: "GradientBased",
		}
	}

	// Sort results by concurrency level
	sortedResults := make([]ConcurrentTestResult, len(results))
	copy(sortedResults, results)

	// Simple bubble sort by concurrency
	for i := 0; i < len(sortedResults)-1; i++ {
		for j := 0; j < len(sortedResults)-i-1; j++ {
			if sortedResults[j].Concurrency > sortedResults[j+1].Concurrency {
				sortedResults[j], sortedResults[j+1] = sortedResults[j+1], sortedResults[j]
			}
		}
	}

	// Find the point where QPS growth rate drops below threshold
	for i := 1; i < len(sortedResults); i++ {
		prev := sortedResults[i-1]
		curr := sortedResults[i]

		// Skip if either result has no successful requests
		if prev.Metrics.SuccessfulRequests == 0 || curr.Metrics.SuccessfulRequests == 0 {
			continue
		}

		// Calculate QPS gradient
		concurrencyDiff := float64(curr.Concurrency - prev.Concurrency)
		if concurrencyDiff == 0 {
			continue
		}

		// Convert analyzer.Float64 to float64
		prevQPS := float64(prev.Metrics.QPS)
		currQPS := float64(curr.Metrics.QPS)

		qpsGradient := (currQPS - prevQPS) / concurrencyDiff

		// If gradient is below threshold, we've found the bottleneck
		if qpsGradient < g.Threshold {
			return &BottleneckResult{
				Concurrency:   prev.Concurrency,
				QPS:           prevQPS,
				IsBottleneck:  true,
				AlgorithmUsed: "GradientBased",
			}
		}
	}

	// If no bottleneck found, return the last result
	lastResult := sortedResults[len(sortedResults)-1]
	return &BottleneckResult{
		Concurrency:   lastResult.Concurrency,
		QPS:           float64(lastResult.Metrics.QPS),
		IsBottleneck:  false,
		AlgorithmUsed: "GradientBased",
	}
}

// StatisticalBasedDetector implements bottleneck detection using statistical method
type StatisticalBasedDetector struct {
	// WindowSize is the size of the sliding window for calculating averages
	WindowSize int

	// Threshold is the coefficient of variation threshold
	Threshold float64
}

// NewStatisticalBasedDetector creates a new statistical-based bottleneck detector
func NewStatisticalBasedDetector(windowSize int, threshold float64) *StatisticalBasedDetector {
	return &StatisticalBasedDetector{
		WindowSize: windowSize,
		Threshold:  threshold,
	}
}

// DetectBottleneck finds the concurrency level at which QPS bottleneck occurs using statistical method
func (s *StatisticalBasedDetector) DetectBottleneck(results []ConcurrentTestResult) *BottleneckResult {
	if len(results) < s.WindowSize {
		return &BottleneckResult{
			IsBottleneck:  false,
			AlgorithmUsed: "StatisticalBased",
		}
	}

	// Sort results by concurrency level
	sortedResults := make([]ConcurrentTestResult, len(results))
	copy(sortedResults, results)

	// Simple bubble sort by concurrency
	for i := 0; i < len(sortedResults)-1; i++ {
		for j := 0; j < len(sortedResults)-i-1; j++ {
			if sortedResults[j].Concurrency > sortedResults[j+1].Concurrency {
				sortedResults[j], sortedResults[j+1] = sortedResults[j+1], sortedResults[j]
			}
		}
	}

	// Calculate QPS values
	qpsValues := make([]float64, len(sortedResults))
	for i, result := range sortedResults {
		qpsValues[i] = float64(result.Metrics.QPS)
	}

	// Find the point where coefficient of variation drops below threshold
	for i := s.WindowSize; i < len(qpsValues); i++ {
		// Calculate mean and standard deviation for the window
		sum := 0.0
		for j := i - s.WindowSize; j < i; j++ {
			sum += qpsValues[j]
		}
		mean := sum / float64(s.WindowSize)

		if mean == 0 {
			continue
		}

		// Calculate standard deviation
		varianceSum := 0.0
		for j := i - s.WindowSize; j < i; j++ {
			diff := qpsValues[j] - mean
			varianceSum += diff * diff
		}
		stdDev := varianceSum / float64(s.WindowSize)
		if stdDev > 0 {
			stdDev = math.Sqrt(stdDev)
		}

		// Calculate coefficient of variation
		cv := stdDev / mean

		// If coefficient of variation is below threshold, we've found the bottleneck
		if cv < s.Threshold {
			return &BottleneckResult{
				Concurrency:   sortedResults[i-s.WindowSize].Concurrency,
				QPS:           float64(sortedResults[i-s.WindowSize].Metrics.QPS),
				IsBottleneck:  true,
				AlgorithmUsed: "StatisticalBased",
			}
		}
	}

	// If no bottleneck found, return the last result
	lastResult := sortedResults[len(sortedResults)-1]
	return &BottleneckResult{
		Concurrency:   lastResult.Concurrency,
		QPS:           float64(lastResult.Metrics.QPS),
		IsBottleneck:  false,
		AlgorithmUsed: "StatisticalBased",
	}
}

// DetectQPSBottleneck is a convenience function that uses default gradient-based detection
func (c *ConcurrentComparison) DetectQPSBottleneck() *BottleneckResult {
	detector := NewGradientBasedDetector(0.05) // 5% threshold
	return detector.DetectBottleneck(c.TestResults)
}
