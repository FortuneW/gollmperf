package reporter

import (
	"math"
	"sort"
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

	// Reason is the reason for the bottleneck
	Reason string

	// QPS is the QPS value at bottleneck
	QPS float64

	// TokensPerSec is the Tokens/Sec value at bottleneck
	TokensPerSec float64

	// AverageLatencyMS is the average latency at bottleneck
	AverageLatencyMS int64

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
			Concurrency:      results[0].Concurrency,
			QPS:              float64(results[0].Metrics.QPS),
			TokensPerSec:     float64(results[0].Metrics.TokensPerSecond),
			AverageLatencyMS: results[0].Metrics.AverageLatency.Milliseconds(),
			IsBottleneck:     false,
			AlgorithmUsed:    "GradientBased",
		}
	}

	// Sort results by concurrency level
	sortedResults := make([]ConcurrentTestResult, len(results))
	copy(sortedResults, results)

	// Sort by concurrency using Go's built-in sort function
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Concurrency < sortedResults[j].Concurrency
	})

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
				Concurrency:      prev.Concurrency,
				QPS:              prevQPS,
				TokensPerSec:     float64(prev.Metrics.TokensPerSecond),
				AverageLatencyMS: prev.Metrics.AverageLatency.Milliseconds(),
				IsBottleneck:     true,
				AlgorithmUsed:    "GradientBased",
			}
		}
	}

	// If no bottleneck found, return the last result
	lastResult := sortedResults[len(sortedResults)-1]
	return &BottleneckResult{
		Concurrency:      lastResult.Concurrency,
		QPS:              float64(lastResult.Metrics.QPS),
		TokensPerSec:     float64(lastResult.Metrics.TokensPerSecond),
		AverageLatencyMS: lastResult.Metrics.AverageLatency.Milliseconds(),
		IsBottleneck:     false,
		AlgorithmUsed:    "GradientBased",
	}
}

// StatisticalBasedDetector implements bottleneck detection using statistical method
type StatisticalBasedDetector struct {
	// WindowSize is the size of the sliding window for calculating averages
	WindowSize int

	// Threshold is the coefficient of variation threshold
	Threshold float64
}

// LatencyBasedDetector implements bottleneck detection using latency ratio method
type LatencyBasedDetector struct {
	// Threshold is the minimum latency growth rate to concurrency growth rate ratio to consider as bottleneck
	Threshold float64
	// UseFirstTokenLatency indicates whether to use first token latency instead of end-to-end latency
	UseFirstTokenLatency bool
}

// NewStatisticalBasedDetector creates a new statistical-based bottleneck detector
func NewStatisticalBasedDetector(windowSize int, threshold float64) *StatisticalBasedDetector {
	return &StatisticalBasedDetector{
		WindowSize: windowSize,
		Threshold:  threshold,
	}
}

// NewLatencyBasedDetector creates a new latency-based bottleneck detector
func NewLatencyBasedDetector(threshold float64) *LatencyBasedDetector {
	return &LatencyBasedDetector{
		Threshold:            threshold,
		UseFirstTokenLatency: false,
	}
}

// NewFirstTokenLatencyBasedDetector creates a new first token latency-based bottleneck detector
func NewFirstTokenLatencyBasedDetector(threshold float64) *LatencyBasedDetector {
	return &LatencyBasedDetector{
		Threshold:            threshold,
		UseFirstTokenLatency: true,
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

	// Sort by concurrency using Go's built-in sort function
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Concurrency < sortedResults[j].Concurrency
	})

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
				Concurrency:      sortedResults[i-s.WindowSize].Concurrency,
				QPS:              float64(sortedResults[i-s.WindowSize].Metrics.QPS),
				TokensPerSec:     float64(sortedResults[i-s.WindowSize].Metrics.TokensPerSecond),
				AverageLatencyMS: sortedResults[i-s.WindowSize].Metrics.AverageLatency.Milliseconds(),
				IsBottleneck:     true,
				AlgorithmUsed:    "StatisticalBased",
			}
		}
	}

	// If no bottleneck found, return the last result
	lastResult := sortedResults[len(sortedResults)-1]
	return &BottleneckResult{
		Concurrency:      lastResult.Concurrency,
		QPS:              float64(lastResult.Metrics.QPS),
		TokensPerSec:     float64(lastResult.Metrics.TokensPerSecond),
		AverageLatencyMS: lastResult.Metrics.AverageLatency.Milliseconds(),
		IsBottleneck:     false,
		AlgorithmUsed:    "StatisticalBased",
	}
}

// DetectBottleneck finds the concurrency level at which latency bottleneck occurs using ratio method
func (l *LatencyBasedDetector) DetectBottleneck(results []ConcurrentTestResult) *BottleneckResult {
	if len(results) == 0 {
		return &BottleneckResult{
			IsBottleneck:  false,
			AlgorithmUsed: "LatencyBased",
		}
	} else if len(results) == 1 {
		// Get the appropriate latency value based on the flag
		var avgLatency int64
		if l.UseFirstTokenLatency {
			avgLatency = results[0].Metrics.AverageFirstTokenLatency.Milliseconds()
		} else {
			avgLatency = results[0].Metrics.AverageLatency.Milliseconds()
		}

		return &BottleneckResult{
			Concurrency:      results[0].Concurrency,
			QPS:              float64(results[0].Metrics.QPS),
			TokensPerSec:     float64(results[0].Metrics.TokensPerSecond),
			AverageLatencyMS: avgLatency,
			IsBottleneck:     false,
			AlgorithmUsed:    "LatencyBased",
		}
	}

	// Sort results by concurrency level
	sortedResults := make([]ConcurrentTestResult, len(results))
	copy(sortedResults, results)

	// Sort by concurrency using Go's built-in sort function
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Concurrency < sortedResults[j].Concurrency
	})

	// Find the point where latency growth rate exceeds concurrency growth rate by threshold
	for i := 1; i < len(sortedResults); i++ {
		prev := sortedResults[i-1]
		curr := sortedResults[i]

		// Skip if either result has no successful requests
		if prev.Metrics.SuccessfulRequests == 0 || curr.Metrics.SuccessfulRequests == 0 {
			continue
		}

		// Get the appropriate latency values based on the flag
		var prevLatency, currLatency float64
		if l.UseFirstTokenLatency {
			prevLatency = float64(prev.Metrics.AverageFirstTokenLatency)
			currLatency = float64(curr.Metrics.AverageFirstTokenLatency)
		} else {
			prevLatency = float64(prev.Metrics.AverageLatency)
			currLatency = float64(curr.Metrics.AverageLatency)
		}

		// Skip if either result has zero average latency
		if prevLatency == 0 || currLatency == 0 {
			continue
		}

		// Calculate concurrency growth rate
		concurrencyDiff := float64(curr.Concurrency - prev.Concurrency)
		if concurrencyDiff == 0 {
			continue
		}

		// Calculate latency growth rate
		latencyDiff := currLatency - prevLatency

		// Calculate growth rate ratio (latency growth rate / concurrency growth rate)
		// We want to detect when latency grows faster than concurrency
		ratio := (latencyDiff / prevLatency) / (concurrencyDiff / float64(prev.Concurrency))

		// If ratio is above threshold, we've found the bottleneck
		if ratio > l.Threshold {
			return &BottleneckResult{
				Concurrency:      prev.Concurrency,
				QPS:              float64(prev.Metrics.QPS),
				TokensPerSec:     float64(prev.Metrics.TokensPerSecond),
				AverageLatencyMS: int64(prevLatency / 1000000),
				IsBottleneck:     true,
				AlgorithmUsed:    "LatencyBased",
			}
		}
	}

	// If no bottleneck found, return the last result
	lastResult := sortedResults[len(sortedResults)-1]

	// Get the appropriate latency value for the last result
	var lastLatency int64
	if l.UseFirstTokenLatency {
		lastLatency = lastResult.Metrics.AverageFirstTokenLatency.Milliseconds()
	} else {
		lastLatency = lastResult.Metrics.AverageLatency.Milliseconds()
	}

	return &BottleneckResult{
		Concurrency:      lastResult.Concurrency,
		QPS:              float64(lastResult.Metrics.QPS),
		TokensPerSec:     float64(lastResult.Metrics.TokensPerSecond),
		AverageLatencyMS: lastLatency,
		IsBottleneck:     false,
		AlgorithmUsed:    "LatencyBased",
	}
}

// DetectQPSBottleneck is a convenience function that uses default gradient-based detection
func (c *ConcurrentComparison) DetectQPSBottleneck() *BottleneckResult {
	detector := NewGradientBasedDetector(0.05) // 5% threshold
	return detector.DetectBottleneck(c.TestResults)
}
