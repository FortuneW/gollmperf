package engine

import (
	"sync"
	"time"

	"github.com/FortuneW/gollmperf/internal/provider"
	"github.com/FortuneW/qlog"
)

var stressLog = qlog.GetRLog("engine.stress")

var onceWarmup sync.Once

func (e *Engine) RunStress(dataset []provider.AnyParams) ([]*Result, error) {
	// Warmup phase
	var err error

	if e.config.Test.Warmup > 0 {
		onceWarmup.Do(func() {
			stressLog.Infof("Starting warmup for %v...", e.config.Test.Warmup)
			err = e.runWarmup(dataset)
		})
		if err != nil {
			return nil, err
		}
	}

	// Actual stress test
	stressLog.Infof("Starting stress testing for %v or %d requests/concurrency with concurrency %d...",
		e.config.Test.Duration, e.config.Test.RequestsPerConcurrency, e.config.Test.Concurrency)

	// Create channel for results
	resultsChan := make(chan *Result, 1000) // Buffered channel to prevent blocking
	var results []*Result
	resultsMutex := sync.Mutex{}

	testDuration := e.config.Test.Duration

	// Start worker goroutines
	concurrency := e.getConcurrency()
	wg := e.startWorkers(concurrency, func(workerID int, wg *sync.WaitGroup) {
		// Each worker runs until either duration is reached or requests per concurrency is met
		workerStartTime := time.Now()
		reqIndex := workerID // Start with different index for each worker
		requestsCompleted := 0
		maxRequests := e.config.Test.RequestsPerConcurrency

		for (testDuration <= 0 || time.Since(workerStartTime) < testDuration) &&
			(maxRequests <= 0 || requestsCompleted < maxRequests) {

			// Get a request from dataset (round-robin)
			req := dataset[reqIndex%len(dataset)]
			reqIndex++

			e.executeWorkerJob(req, resultsChan)
			requestsCompleted++

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
