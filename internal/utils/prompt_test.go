package utils

import (
	"os"
	"testing"

	"github.com/FortuneW/gollmperf/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetSystemPrompt(t *testing.T) {
	// Test direct content
	cfg1 := &config.SystemPromptTemplate{
		Content: "You are a helpful assistant.",
		Path:    "./test_prompt.txt",
		Enable:  true,
	}
	content1 := GetSystemPrompt(cfg1)
	assert.Equal(t, "You are a helpful assistant.", content1)

	// Test file path
	// Create temporary test file
	tempFile, err := os.CreateTemp(".", "test_prompt_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write test content
	testContent := "This is a test system prompt from file."
	_, err = tempFile.WriteString(testContent)
	assert.NoError(t, err)
	tempFile.Close()

	// Test reading from file
	cfg2 := &config.SystemPromptTemplate{
		Content: "",
		Path:    tempFile.Name(),
		Enable:  true,
	}
	content2 := GetSystemPrompt(cfg2)
	assert.Equal(t, testContent, content2)

	// Test both empty
	cfg3 := &config.SystemPromptTemplate{
		Content: "",
		Path:    "",
		Enable:  true,
	}
	content3 := GetSystemPrompt(cfg3)
	assert.Equal(t, "", content3)

	// Test file not found
	cfg4 := &config.SystemPromptTemplate{
		Content: "",
		Path:    "./non_existent_file.txt",
		Enable:  true,
	}
	content4 := GetSystemPrompt(cfg4)
	assert.Equal(t, "", content4)
}
