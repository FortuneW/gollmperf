package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/FortuneW/gollmperf/internal/provider"
)

var bufferPool = sync.Pool{
	New: func() any {
		const maxCapacity = 1 * 1024 * 1024
		buf := make([]byte, maxCapacity)
		return buf
	},
}

func getBuffer() []byte {
	return bufferPool.Get().([]byte)
}

func putBuffer(buf []byte) {
	buf = buf[:0]
	bufferPool.Put(buf)
}

// addSystemPromptToMessages adds system prompt to the beginning of messages array
func addSystemPromptToMessages(reqCase provider.AnyParams, systemPrompt string) provider.AnyParams {
	// Add system prompt to the beginning of messages array if it exists
	if strings.TrimSpace(systemPrompt) != "" {
		if messages, ok := reqCase["messages"].([]interface{}); ok {
			// Handle interface type messages
			// Check if the first message is already a system message
			if len(messages) > 0 {
				if msgMap, ok := messages[0].(map[string]interface{}); ok {
					if role, ok := msgMap["role"].(string); ok && role == "system" {
						// Replace existing system message
						msgMap["content"] = systemPrompt
						messages[0] = msgMap
						reqCase["messages"] = messages
						return reqCase
					}
				}
			}
			// Create a new messages array including system prompt and original messages
			newMessages := make([]interface{}, len(messages)+1)
			newMessages[0] = map[string]interface{}{
				"role":    "system",
				"content": systemPrompt,
			}
			copy(newMessages[1:], messages)
			reqCase["messages"] = newMessages
		}
	}
	return reqCase
}

// LoadDataset loads test data from a file
func LoadDataset(filePath, fileType string, systemPrompt string) ([]provider.AnyParams, error) {
	switch fileType {
	case "jsonl":
		requests, err := loadJSONLDataset(filePath)
		if err != nil {
			return nil, err
		}
		// Add system prompt to all requests if provided
		if strings.TrimSpace(systemPrompt) != "" {
			for i := range requests {
				requests[i] = addSystemPromptToMessages(requests[i], systemPrompt)
			}
		}
		return requests, nil
	default:
		return nil, fmt.Errorf("unsupported dataset type: %s", fileType)
	}
}

// loadJSONLDataset loads dataset from JSONL file
func loadJSONLDataset(filePath string) ([]provider.AnyParams, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var requests []provider.AnyParams
	scanner := bufio.NewScanner(file)

	buf := getBuffer()
	defer putBuffer(buf)

	// 设置 scanner 的缓冲区
	scanner.Buffer(buf, cap(buf))

	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse JSON line
		var jsonReq provider.AnyParams

		if err := json.Unmarshal([]byte(line), &jsonReq); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNumber, err)
		}

		requests = append(requests, jsonReq)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return requests, nil
}
