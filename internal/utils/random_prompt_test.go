package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// mockTokenizeServer creates a mock vLLM tokenize server for testing
func mockTokenizeServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tokenize" {
			t.Errorf("Expected path /tokenize, got %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TokenizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Simple token count simulation: approximately 1 token per 3 characters
		// This is a rough approximation for testing purposes
		count := len(req.Prompt) / 3
		if count < 1 {
			count = 1
		}

		resp := TokenizeResponse{
			Count:       count,
			MaxModelLen: 262144,
			Tokens:      make([]int, count),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestBuildTokenizeURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
		wantErr  bool
	}{
		{
			name:     "full URL with path",
			endpoint: "http://localhost:63535/v1/chat/completions",
			want:     "http://localhost:63535/tokenize",
			wantErr:  false,
		},
		{
			name:     "URL without path",
			endpoint: "http://localhost:63535",
			want:     "http://localhost:63535/tokenize",
			wantErr:  false,
		},
		{
			name:     "host:port only",
			endpoint: "localhost:63535",
			want:     "http://localhost:63535/tokenize",
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "https URL",
			endpoint: "https://api.example.com/v1/chat/completions",
			want:     "http://api.example.com/tokenize",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildTokenizeURL(tt.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildTokenizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildTokenizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCallTokenizeAPI(t *testing.T) {
	server := mockTokenizeServer(t)
	defer server.Close()

	tests := []struct {
		name    string
		prompt  string
		wantMin int
		wantErr bool
	}{
		{
			name:    "short prompt",
			prompt:  "hello",
			wantMin: 1,
			wantErr: false,
		},
		{
			name:    "long prompt",
			prompt:  "this is a longer prompt with more characters",
			wantMin: 5,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := CallTokenizeAPI(server.URL+"/tokenize", tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("callTokenizeAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if count < tt.wantMin {
				t.Errorf("callTokenizeAPI() count = %v, want at least %v", count, tt.wantMin)
			}
		})
	}
}

func TestCallTokenizeAPI_InvalidServer(t *testing.T) {
	// Test with invalid server URL
	_, err := CallTokenizeAPI("http://invalid-server:99999/tokenize", "test")
	if err == nil {
		t.Error("callTokenizeAPI() expected error for invalid server")
	}
}

func TestGenerateRandomWords(t *testing.T) {
	tests := []struct {
		name         string
		targetTokens int
		wantNonEmpty bool
	}{
		{
			name:         "small token count",
			targetTokens: 10,
			wantNonEmpty: true,
		},
		{
			name:         "medium token count",
			targetTokens: 100,
			wantNonEmpty: true,
		},
		{
			name:         "large token count",
			targetTokens: 1000,
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateRandomWords(tt.targetTokens)
			if tt.wantNonEmpty && got == "" {
				t.Error("generateRandomWords() returned empty string")
			}
			// Check that we got some words (space-separated)
			words := strings.Fields(got)
			if len(words) == 0 {
				t.Error("generateRandomWords() returned no words")
			}
			t.Logf("Generated %d words for target %d tokens", len(words), tt.targetTokens)
		})
	}
}

func TestGetRandomPromptByTokenCount(t *testing.T) {
	server := mockTokenizeServer(t)
	defer server.Close()

	tests := []struct {
		name         string
		endpoint     string
		targetTokens int
		wantErr      bool
	}{
		{
			name:         "small token count",
			endpoint:     server.URL + "/v1/chat/completions",
			targetTokens: 5,
			wantErr:      false,
		},
		{
			name:         "larger token count",
			endpoint:     server.URL + "/v1/chat/completions",
			targetTokens: 20,
			wantErr:      false,
		},
		{
			name:         "empty endpoint",
			endpoint:     "",
			targetTokens: 10,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRandomPromptByTokenCount(tt.endpoint, tt.targetTokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRandomPromptByTokenCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("GetRandomPromptByTokenCount() returned empty string for valid input")
			}
		})
	}
}

func TestGetRandomPromptByTokenCount_MaxRetries(t *testing.T) {
	// Create a server that always returns a fixed count to trigger max retry
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenizeResponse{
			Count:       100, // Always return 100, making it impossible to match target
			MaxModelLen: 262144,
			Tokens:      make([]int, 100),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Request 50 tokens, but server always returns 100
	// Function should use word-based estimation to truncate and return success
	prompt, err := GetRandomPromptByTokenCount(server.URL+"/v1/chat/completions", 50)
	if err != nil {
		t.Errorf("GetRandomPromptByTokenCount() unexpected error: %v", err)
	}
	if prompt == "" {
		t.Error("GetRandomPromptByTokenCount() returned empty prompt")
	}
	// Verify the prompt was truncated (should have around 25 words for 50 tokens target)
	words := strings.Fields(prompt)
	if len(words) > 30 {
		t.Errorf("Expected prompt to be truncated to ~25 words, got %d words", len(words))
	}
}

func TestGetRandomPromptByTokenCount_RealAPI(t *testing.T) {
	endpoint := os.Getenv("LLM_API_ENDPOINT")
	if endpoint == "" {
		t.Skip("LLM_API_ENDPOINT environment variable not set, skipping real API test")
	}
	str, err := GetRandomPromptByTokenCount(endpoint, 10000)
	if err == nil {
		// t.Logf("GetRandomPromptByTokenCount() success: %s", str)
		_ = str
	} else {
		t.Errorf("GetRandomPromptByTokenCount() error: %v", err)
	}
}
