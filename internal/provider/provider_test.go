package provider

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

var (
	apiKey = func() (key string) {
		keys := []string{"LLM_API_KEY", "OPENAI_API_KEY", "DASHSCOPE_API_KEY"}
		for _, k := range keys {
			if key = os.Getenv(k); key != "" {
				return
			}
		}
		return
	}()
)

func TestOpenAIProvider_OpenAI(t *testing.T) {
	if apiKey == "" {
		t.Skip("Skipping test: no API key available")
	}
	provider := NewOpenAIProvider(apiKey, "", time.Second*10)
	anyParams := AnyParams{
		"messages": []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	priorityParam := AnyParams{
		"model":      "gpt-3.5-turbo",
		"stream":     true,
		"max_tokens": 100,
	}

	resp, err := provider.SendRequest(priorityParam, anyParams, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", resp)
}

func TestOpenAIProvider_Qwen(t *testing.T) {
	if apiKey == "" {
		t.Skip("Skipping test: no API key available")
	}
	provider := NewQwenProvider(apiKey, "", time.Second*10)

	anyParams := AnyParams{
		"messages": []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	priorityParam := AnyParams{
		"model":  "qwen-plus",
		"stream": false,
		"extra_body": map[string]any{
			"enable_thinking": false,
		},
	}

	resp, err := provider.SendRequest(priorityParam, anyParams, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Verify response has content
	if len(resp.Choices) == 0 {
		t.Error("Expected at least one choice in response")
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok || content == "" {
		t.Error("Expected content in response message")
	}

	t.Logf("%+v", resp)
}

func TestOpenAIProvider_Streaming(t *testing.T) {
	if apiKey == "" {
		t.Skip("Skipping test: no API key available")
	}
	provider := NewQwenProvider(apiKey, "", time.Second*10)

	anyParams := AnyParams{
		"messages": []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	priorityParam := AnyParams{
		"model":  "qwen-plus",
		"stream": true,
		"extra_body": map[string]any{
			"enable_thinking": false,
		},
	}

	resp, err := provider.SendRequest(priorityParam, anyParams, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Verify streaming response has content
	if len(resp.Choices) == 0 {
		t.Fatal("Expected at least one choice in response")
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok || content == "" {
		t.Error("Expected content in response message")
	}

	// Verify timing information
	if resp.Latency == 0 {
		t.Error("Expected latency to be set")
	}

	// For streaming, first token latency should be different from total latency
	// (though in our implementation they might be close)
	if resp.FirstTokenLatency == 0 {
		t.Error("Expected first token latency to be set for streaming response")
	}

	b, _ := json.MarshalIndent(resp, "", " ")
	t.Logf("Response: %s", b)
	t.Logf("Content length: %d characters", len(content))
	t.Logf("Latency: %v, First token latency: %v", resp.Latency, resp.FirstTokenLatency)
}
