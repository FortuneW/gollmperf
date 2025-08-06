package collector

import (
	"time"

	"github.com/FortuneW/gollmperf/internal/engine"
)

// Collector collects and stores test results
type Collector struct {
	results []*engine.Result
}

// NewCollector creates a new collector
func NewCollector(results []*engine.Result) *Collector {
	if results == nil {
		results = make([]*engine.Result, 0)
	}
	return &Collector{
		results: results,
	}
}

// AddResult adds a result to the collector
func (c *Collector) AddResult(result *engine.Result) {
	c.results = append(c.results, result)
}

// GetAllResults returns all collected results
func (c *Collector) GetAllResults() []*engine.Result {
	return c.results
}

// GetSuccessfulResults returns only successful results
func (c *Collector) GetSuccessfulResults() []*engine.Result {
	successful := make([]*engine.Result, 0)
	for _, result := range c.results {
		if result.Success {
			successful = append(successful, result)
		}
	}
	return successful
}

// GetFailedResults returns only failed results
func (c *Collector) GetFailedResults() []*engine.Result {
	failed := make([]*engine.Result, 0)
	for _, result := range c.results {
		if !result.Success {
			failed = append(failed, result)
		}
	}
	return failed
}

// GetTotalCount returns the total number of results
func (c *Collector) GetTotalCount() int {
	return len(c.results)
}

// GetSuccessCount returns the number of successful results
func (c *Collector) GetSuccessCount() int {
	count := 0
	for _, result := range c.results {
		if result.Success {
			count++
		}
	}
	return count
}

// GetFailureCount returns the number of failed results
func (c *Collector) GetFailureCount() int {
	return c.GetTotalCount() - c.GetSuccessCount()
}

// GetTestDuration returns the duration from first to last result
func (c *Collector) GetTestDuration() time.Duration {
	if len(c.results) == 0 {
		return 0
	}

	first := c.results[0].StartTime
	last := c.results[0].EndTime

	for _, result := range c.results {
		if result.StartTime.Before(first) {
			first = result.StartTime
		}
		if result.EndTime.After(last) {
			last = result.EndTime
		}
	}

	return last.Sub(first)
}
