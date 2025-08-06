package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/user/llmperf/internal/provider"
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

// LoadDataset loads test data from a file
func LoadDataset(filePath, fileType string) ([]provider.AnyParams, error) {
	switch fileType {
	case "jsonl":
		return loadJSONLDataset(filePath)
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

// CreateDefaultDataset creates a default dataset for testing
func CreateDefaultDataset() []provider.AnyParams {
	return []provider.AnyParams{
		{
			"messages": []provider.Message{
				{Role: "user", Content: "Write a short poem about programming."},
			},
		},
		{
			"messages": []provider.Message{
				{Role: "user", Content: "Explain what is a neural network in simple terms."},
			},
		},
		{
			"messages": []provider.Message{
				{Role: "user", Content: "How to optimize a Python function for better performance?"},
			},
		},
	}
}
