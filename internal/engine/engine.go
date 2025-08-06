package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/FortuneW/qlog"
	"github.com/user/llmperf/internal/config"
	"github.com/user/llmperf/internal/provider"
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
	Error             string             `json:"error,omitempty"`
	ErrorType         string             `json:"error_type,omitempty"` // New field for error categorization
	StartTime         time.Time          `json:"start_time"`
	EndTime           time.Time          `json:"end_time"`
	RefResponse       *provider.Response `json:"-"`
}

var mlog = qlog.GetRLog("engine")

// categorizeError categorizes errors into network errors or other errors
func categorizeError(err error) string {
	// Check for network-related errors
	if ok, str := isNetworkError(err); ok {
		return str
	}

	// Default to other errors
	return err.Error()
}

// isNetworkError checks if an error is network-related
func isNetworkError(err error) (bool, string) {
	// Check for common network error types
	// Note: We can't directly import "net" in this file as it's already imported
	// We'll check the error string for network-related keywords
	errStr := err.Error()

	// Common network error indicators
	networkIndicators := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"dial tcp",
		"network is unreachable",
		"no such host",
		"i/o timeout",
		"context deadline exceeded",
	}

	for _, indicator := range networkIndicators {
		if strings.Contains(strings.ToLower(errStr), indicator) {
			return true, indicator
		}
	}

	return false, ""
}

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
			reqIndex := workerID // 每个worker起始索引不同，避免请求重复

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
		result.Error = err.Error()
		result.ErrorType = categorizeError(err)
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
