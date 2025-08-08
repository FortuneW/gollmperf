package provider

import "time"

// QwenProvider implements the Provider interface for Qwen(same as OpenAI)
type QwenProvider struct {
	oai *OpenAIProvider
}

// NewQwenProvider creates a new QwenProvider
func NewQwenProvider(apiKey, endpoint, model string, timeout time.Duration) *QwenProvider {
	if endpoint == "" {
		endpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
		mlog.Infof("Created Qwen provider [%s] with model [%s]", endpoint, model)
	}
	return &QwenProvider{
		oai: NewOpenAIProvider(apiKey, endpoint, model, timeout),
	}
}

// Name returns the provider name
func (p *QwenProvider) Name() string {
	return "qwen"
}

// SendRequest sends a request to Qwen API
func (p *QwenProvider) SendRequest(priorityParams, anyParam AnyParams, headers map[string]string) (*Response, *Error) {
	return p.oai.SendRequest(priorityParams, anyParam, headers)
}

// SupportsStreaming returns whether Qwen supports streaming
func (p *QwenProvider) SupportsStreaming() bool {
	return p.oai.SupportsStreaming()
}
