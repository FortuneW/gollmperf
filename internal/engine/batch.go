package engine

import (
	"sync"

	"github.com/FortuneW/gollmperf/internal/provider"
	"github.com/FortuneW/qlog"
)

var batchLog = qlog.GetRLog("engine.batch")

// RunBatch runs a batch test
func (e *Engine) RunBatch(dataset []provider.AnyParams) ([]*Result, error) {
	batchLog.Infof("Starting batch testing with concurrency %d...", e.config.Test.Concurrency)

	// Create results slice with exact capacity
	results := make([]*Result, len(dataset))

	// Channel to collect results with their indices
	resultsChan := make(chan workerResult, len(dataset))

	// Create jobs channel
	jobsChan := make(chan struct {
		index int
		req   provider.AnyParams
	}, len(dataset))

	// Send all jobs to the jobs channel
	for i, req := range dataset {
		jobsChan <- struct {
			index int
			req   provider.AnyParams
		}{index: i, req: req}
	}
	close(jobsChan)

	// Start worker goroutines
	concurrency := e.getConcurrency()
	wg := e.startWorkers(concurrency, func(workerID int, wg *sync.WaitGroup) {
		// Process jobs from the jobs channel
		for job := range jobsChan {
			result := e.executeRequest(job.req)

			// Send indexed result to results channel
			resultsChan <- workerResult{
				index:  job.index,
				result: result,
			}
		}
	})

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from channel in order
	for indexedRes := range resultsChan {
		results[indexedRes.index] = indexedRes.result
	}

	return results, nil
}
