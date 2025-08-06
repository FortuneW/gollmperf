package engine

import (
	"sync"

	"github.com/FortuneW/gollmperf/internal/provider"
)

// workerResult represents a result from a worker with its index
type workerResult struct {
	index  int
	result *Result
}

// executeWorkerJob executes a single job and sends the result to the results channel
func (e *Engine) executeWorkerJob(job provider.AnyParams, resultsChan chan *Result) {
	result := e.executeRequest(job)

	// Send result to channel (non-blocking)
	select {
	case resultsChan <- result:
	default:
		// If channel is full, skip result to prevent blocking
		mlog.Warn("Result channel full, dropping result")
	}
}

// startWorkers starts a specified number of worker goroutines
func (e *Engine) startWorkers(concurrency int, workerFunc func(workerID int, wg *sync.WaitGroup)) *sync.WaitGroup {
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerFunc(workerID, &wg)
		}(i)
	}

	return &wg
}

// getConcurrency returns the concurrency level, ensuring it's at least 1
func (e *Engine) getConcurrency() int {
	concurrency := e.config.Test.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	return concurrency
}
