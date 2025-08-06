package engine

import (
	"sync"
	"time"

	"github.com/FortuneW/gollmperf/internal/provider"
	"github.com/FortuneW/qlog"
)

var stressLog = qlog.GetRLog("engine.stress")

func (e *Engine) RunStress(dataset []provider.AnyParams) ([]*Result, error) {
	// Warmup phase
	if e.config.Test.Warmup > 0 {
		stressLog.Infof("Starting warmup for %v...", e.config.Test.Warmup)
		if err := e.runWarmup(dataset); err != nil {
			return nil, err
		}
	}

	// Actual stress test
	stressLog.Infof("Starting stress test for %v with concurrency %d...", e.config.Test.Duration, e.config.Test.Concurrency)

	// Create channel for results
	resultsChan := make(chan *Result, 1000) // Buffered channel to prevent blocking
	var results []*Result
	resultsMutex := sync.Mutex{}

	testDuration := e.config.Test.Duration

	// Start worker goroutines
	concurrency := e.getConcurrency()
	wg := e.startWorkers(concurrency, func(workerID int, wg *sync.WaitGroup) {
		// Each worker runs for the specified duration
		workerStartTime := time.Now()
		reqIndex := workerID // Start with different index for each worker

		for time.Since(workerStartTime) < testDuration {
			// Get a request from dataset (round-robin)
			req := dataset[reqIndex%len(dataset)]
			reqIndex++

			e.executeWorkerJob(req, resultsChan)

			// Small delay to prevent overwhelming the system
			time.Sleep(10 * time.Millisecond)
		}
	})

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from channel
	for result := range resultsChan {
		resultsMutex.Lock()
		results = append(results, result)
		resultsMutex.Unlock()
	}

	return results, nil
}
