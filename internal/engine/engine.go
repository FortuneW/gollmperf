package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/FortuneW/gollmperf/internal/config"
	"github.com/FortuneW/gollmperf/internal/provider"
	"github.com/FortuneW/qlog"
)

// Engine is the main test engine
type Engine struct {
	config   *config.Config
	provider provider.Provider
}

// Result represents a single test result
type Result struct {
	RequestTokens     int                `json:"request_tokens"`
	ResponseTokens    int                `json:"response_tokens"`
	Latency           time.Duration      `json:"latency"`
	FirstTokenLatency time.Duration      `json:"first_token_latency,omitempty"`
	Success           bool               `json:"success"`
	Error             *provider.Error    `json:"error,omitempty"`
	StartTime         time.Time          `json:"start_time"`
	EndTime           time.Time          `json:"end_time"`
	RefResponse       *provider.Response `json:"-"`
}

var mlog = qlog.GetRLog("engine")

// NewEngine creates a new test engine
func NewEngine(cfg *config.Config, prov provider.Provider) *Engine {
	if cfg.Model.ParamsTemplate == nil {
		cfg.Model.ParamsTemplate = make(map[string]interface{})
	}
	// override model name
	if len(cfg.Model.Name) > 0 {
		cfg.Model.ParamsTemplate["model"] = cfg.Model.Name
	}
	return &Engine{
		config:   cfg,
		provider: prov,
	}
}

// runWarmup runs the warmup phase
func (e *Engine) runWarmup(dataset []provider.AnyParams) (err error) {
	warmupDuration := e.config.Test.Warmup
	if warmupDuration <= 0 {
		return
	}

	if len(dataset) == 0 {
		return fmt.Errorf("warmup dataset is empty")
	}

	var wg sync.WaitGroup

	for i := 0; i < e.config.Test.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			startTime := time.Now()
			reqIndex := workerID // Each worker has a different starting index to avoid request repetition

			for time.Since(startTime) < warmupDuration {
				req := dataset[reqIndex%len(dataset)]
				res := e.executeRequest(req)
				if !res.Success {
					if err == nil {
						err = fmt.Errorf("warmup failed, first err: %s", res.Error)
					}
					break
				}
				reqIndex++
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	return
}

// executeRequest executes a single request
func (e *Engine) executeRequest(reqCase provider.AnyParams) *Result {
	result := &Result{
		StartTime: time.Now(),
	}

	resp, err := e.provider.SendRequest(e.config.Model.ParamsTemplate, reqCase, e.config.Model.Headers)
	if err != nil {
		// mlog.Warnf("recv api err: %v", err)
		result.Error = err
		result.Success = false
		return result
	}

	result.RefResponse = resp
	result.RequestTokens = resp.Usage.PromptTokens
	result.ResponseTokens = resp.Usage.CompletionTokens
	result.Latency = resp.Latency
	result.FirstTokenLatency = resp.FirstTokenLatency
	result.Success = true
	result.EndTime = time.Now()

	return result
}
