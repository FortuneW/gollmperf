package provider

import (
	"encoding/json"
	"time"

	"github.com/FortuneW/qlog"
)

// Provider interface for different LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// SendRequest sends a request to the LLM and returns the response
	SendRequest(priorityParams AnyParams, anyParam AnyParams, headers map[string]string) (*Response, error)

	// SupportsStreaming returns whether the provider supports streaming
	SupportsStreaming() bool
}

type AnyParams map[string]any

// Message represents a single message in the conversation
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// Response represents an LLM response
type Response struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices,omitempty"`
	Usage   Usage    `json:"usage"`
	// local fields
	Latency           time.Duration `json:"-"`
	FirstTokenLatency time.Duration `json:"-"` // Streaming specific fields

	JsonData string `json:"-"`
}

func (r *Response) String() string {
	if len(r.JsonData) > 0 {
		return r.JsonData
	}
	b, _ := json.Marshal(r)
	return string(b)
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`

	// for stream
	Delta *struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"delta,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

var mlog = qlog.GetRLog("API")
