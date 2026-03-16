package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

const (
	maxRetryCount    = 10
	avgCharsPerToken = 4 // 英文单词平均约 4 个字符对应 1 个 token（包含空格）
)

// TokenizeRequest represents the request body for /tokenize endpoint
type TokenizeRequest struct {
	Prompt string `json:"prompt"`
}

// TokenizeResponse represents the response from /tokenize endpoint
type TokenizeResponse struct {
	Count       int      `json:"count"`
	MaxModelLen int      `json:"max_model_len"`
	Tokens      []int    `json:"tokens"`
	TokenStrs   []string `json:"token_strs"`
}

// GetRandomPromptByTokenCount generates a random prompt with specified token count
// It calls the vLLM /tokenize endpoint to verify token count and adjusts accordingly
func GetRandomPromptByTokenCount(endpoint string, targetTokens int) (string, error) {
	// Parse endpoint to extract host:port
	tokenizeURL, err := BuildTokenizeURL(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to build tokenize URL: %w", err)
	}

	// Generate initial random words based on target token count
	prompt := generateRandomWords(targetTokens)

	// Loop to adjust prompt to match target token count
	for i := 0; i < maxRetryCount; i++ {
		count, err := CallTokenizeAPI(tokenizeURL, prompt)
		if err != nil {
			return "", fmt.Errorf("tokenize API call failed: %w", err)
		}

		mlog.Infof("Random Token count: %d, Target tokens: %d, Prompt length: %d", count, targetTokens, len(prompt))

		// Allow small tolerance (within 2% or 10 tokens, whichever is larger)
		tolerance := max(10, targetTokens/50)
		if abs(count-targetTokens) <= tolerance {
			return prompt, nil
		}

		if count < targetTokens {
			// Need more tokens, calculate how many chars to add based on measured ratio
			charsPerToken := float64(len(prompt)) / float64(count)
			neededTokens := targetTokens - count
			// Add 80% of estimated needed chars to avoid overshooting
			charsToAdd := int(float64(neededTokens) * charsPerToken * 0.8)
			if charsToAdd < 10 {
				charsToAdd = 10
			}
			// Generate words that approximately match the needed chars
			extraWords := int(float64(charsToAdd) / 6) // ~6 chars per word including space
			if extraWords < 1 {
				extraWords = 1
			}
			extra := generateRandomWords(extraWords)
			prompt = prompt + " " + extra
		} else {
			// Too many tokens, calculate current ratio and truncate proportionally
			// Use measured ratio: len(prompt)/count gives chars per token
			charsPerToken := float64(len(prompt)) / float64(count)
			// Target chars = targetTokens * charsPerToken * 0.9 (safety factor)
			targetChars := int(float64(targetTokens) * charsPerToken * 0.9)
			if targetChars < 1 {
				targetChars = 1
			}
			if targetChars < len(prompt) {
				prompt = prompt[:targetChars]
				// Trim to last complete word to avoid partial words
				if idx := strings.LastIndex(prompt, " "); idx > 0 {
					prompt = prompt[:idx]
				}
			}
		}
	}

	// After max retries, if still over target, use simple estimation to truncate
	// Simple estimation: 1 word ≈ 1 token, 1 space ≈ 1 token
	words := strings.Fields(prompt)
	currentTokens := len(words) + len(words) - 1 // words + spaces between them

	if currentTokens > targetTokens {
		// Need to remove words to fit within target
		// Each word removal reduces tokens by ~2 (word + space)
		wordsToKeep := targetTokens / 2
		if wordsToKeep < 1 {
			wordsToKeep = 1
		}
		if wordsToKeep > len(words) {
			wordsToKeep = len(words)
		}
		prompt = strings.Join(words[:wordsToKeep], " ")
	}

	return prompt, nil
}

// abs returns the absolute value of an int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// BuildTokenizeURL parses the endpoint and constructs the tokenize URL
func BuildTokenizeURL(endpoint string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is empty")
	}

	// Parse the endpoint URL
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	// Extract host:port
	host := u.Host
	if host == "" {
		// Try to parse as host:port directly
		host = endpoint
		if strings.HasPrefix(host, "http://") {
			host = strings.TrimPrefix(host, "http://")
		} else if strings.HasPrefix(host, "https://") {
			host = strings.TrimPrefix(host, "https://")
		}
		// Remove path if present
		if idx := strings.Index(host, "/"); idx != -1 {
			host = host[:idx]
		}
	}

	return fmt.Sprintf("http://%s/tokenize", host), nil
}

// CallTokenizeAPI calls the vLLM tokenize endpoint and returns the token count
func CallTokenizeAPI(tokenizeURL, prompt string) (int, error) {
	reqBody := TokenizeRequest{Prompt: prompt}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(tokenizeURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("tokenize API returned status %d", resp.StatusCode)
	}

	var tokenizeResp TokenizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenizeResp); err != nil {
		return 0, err
	}

	return tokenizeResp.Count, nil
}

// generateRandomWords generates a string of random English words
// The targetTokenCount is used to estimate how many words to generate
func generateRandomWords(targetTokenCount int) string {
	// Estimate: each word is about 1-1.5 tokens on average (including space/punctuation)
	// So we generate slightly more words than target tokens
	wordCount := int(float64(targetTokenCount) * 1.2)
	if wordCount < 1 {
		wordCount = 1
	}

	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = gofakeit.Word()
	}

	return strings.Join(words, " ")
}
