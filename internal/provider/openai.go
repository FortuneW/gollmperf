package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	debugRequest  = os.Getenv("DEBUG_LLM_REQUEST")
	debugResponse = os.Getenv("DEBUG_LLM_RESPONSE")
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// NewOpenAIProvider creates a new OpenAIProvider
func NewOpenAIProvider(apiKey, endpoint, model string, timeout time.Duration) *OpenAIProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
		mlog.Infof("Created OpenAI provider [%s] with model [%s]", endpoint, model)
	}

	return &OpenAIProvider{
		apiKey:   apiKey,
		endpoint: endpoint,
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 3 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// mergeRequest merges the request object and additional parameters, returns the merged JSON byte slice,
// and updates the Stream field of the request object.
// Parameters:
//   - priorityParams: The priority request params.
//   - anyParam: Additional parameters.
//
// Returns:
//
//	The merged JSON byte slice.
func (p *OpenAIProvider) mergeRequest(priorityParams, anyParam AnyParams) (data []byte, isStream bool) {
	body := make(map[string]any)

	maps.Copy(body, anyParam)

	// Merge the priority parameter (it will overwrite keys with the same name)
	maps.Copy(body, priorityParams)

	if _, ok := body["stream"]; ok {
		isStream, _ = body["stream"].(bool)
	}

	data, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Errorf("failed to marshal request body: %w", err))
	}

	return
}

// SendRequest sends a request to OpenAI API
func (p *OpenAIProvider) SendRequest(priorityParams, anyParam AnyParams, headers map[string]string) (resp *Response, err error) {
	// Cook request body
	data, isStream := p.mergeRequest(priorityParams, anyParam)

	// debug request body
	if debugRequest == "1" {
		mlog.WithTraceId(fmt.Sprintf("%p", anyParam)).Debugf("Request: %s", string(data))
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// Record start time
	startTime := time.Now()

	// Execute request
	respHttp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer respHttp.Body.Close()

	// Check status code
	if respHttp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respHttp.Body)
		return nil, fmt.Errorf("code %d: %s", respHttp.StatusCode, string(body))
	}

	defer func() {
		// debug response body
		if debugResponse == "1" {
			mlog.WithTraceId(fmt.Sprintf("%p", anyParam)).Debugf("Response: %+v", resp)
		}
	}()

	if isStream {
		resp, err = p.handleStreamingResponse(respHttp, startTime)
		return
	} else {
		resp, err = p.handleNoStreamResponse(respHttp, startTime)
		return
	}
}

// handleNoStreamResponse processes a non-streaming response from OpenAI API
func (p *OpenAIProvider) handleNoStreamResponse(resp *http.Response, startTime time.Time) (*Response, error) {

	// Handle non-streaming response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	response := Response{}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response.Latency = time.Since(startTime)
	if response.FirstTokenLatency == 0 {
		response.FirstTokenLatency = response.Latency // unstreaming same as e2e latency
	}
	return &response, nil
}

// handleStreamingResponse processes a streaming response from OpenAI API
func (p *OpenAIProvider) handleStreamingResponse(resp *http.Response, startTime time.Time) (*Response, error) {
	var (
		response          Response
		firstTokenLatency time.Duration
		content           strings.Builder
		role              string
		finishReason      string
	)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// SSE data lines start with "data: "
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract the JSON data
		data := strings.TrimPrefix(line, "data: ")

		// Check for end of stream
		if data == "[DONE]" {
			continue
		}

		// Parse the SSE event
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			mlog.Errorf("Error unmarshaling response: %v", err)
			continue
		}

		// Record first token latency on first chunk
		if firstTokenLatency == 0 {
			firstTokenLatency = time.Since(startTime)
		}

		// Process choices
		for _, choice := range response.Choices {
			if choice.Delta == nil {
				continue
			}
			content.WriteString(choice.Delta.Content)
			if len(role) == 0 {
				role = choice.Delta.Role
			}
			if len(finishReason) == 0 {
				finishReason = choice.FinishReason
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading streaming response: %w", err)
	}

	// Set the accumulated content as the message
	response.Choices = []Choice{
		{
			FinishReason: finishReason,
			Message: Message{
				Content: content.String(),
				Role:    role,
			},
		},
	}

	// Set timing information
	response.Latency = time.Since(startTime)
	response.FirstTokenLatency = firstTokenLatency

	return &response, nil
}

// SupportsStreaming returns whether OpenAI supports streaming
func (p *OpenAIProvider) SupportsStreaming() bool {
	return true
}
