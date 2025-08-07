package utils

import (
	"os"
	"strings"

	"github.com/FortuneW/gollmperf/internal/config"
	"github.com/FortuneW/qlog"
)

var mlog = qlog.GetRLog("utils")

// GetSystemPrompt retrieves system prompt content based on configuration
// If both content and path are set, content takes precedence
func GetSystemPrompt(cfg *config.SystemPromptTemplate) string {
	// If enable is false, return empty string
	if !cfg.Enable {
		return ""
	}

	// If content is set, return directly
	if strings.TrimSpace(cfg.Content) != "" {
		mlog.Debugf("Using system prompt from content field, length: %d", len(cfg.Content))
		return cfg.Content
	}

	// If path is set, read content from file
	if cfg.Path != "" {
		mlog.Debugf("Reading system prompt from file: %s", cfg.Path)
		content, err := os.ReadFile(cfg.Path)
		if err != nil {
			mlog.Errorf("Failed to read system prompt file: %s, error: %v", cfg.Path, err)
			return ""
		}
		mlog.Debugf("Successfully read system prompt from file, length: %d", len(content))
		return string(content)
	}

	return ""
}
