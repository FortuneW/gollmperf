package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FortuneW/gollmperf/internal/config"
	"github.com/FortuneW/gollmperf/internal/provider"
	"github.com/FortuneW/gollmperf/internal/utils"
)

// TestContext holds the context for running tests
type TestContext struct {
	Config   *config.Config
	Provider provider.Provider
	Dataset  []provider.AnyParams
}

// InitializeTest initializes the test environment based on command line flags and config
func InitializeTest(flags *RunFlags) *TestContext {
	// Must specify config file
	if flags.ConfigPath == "" {
		mlog.Error("Config file must be specified with --config (-c) flag")
		os.Exit(1)
	}

	switch {
	case flags.ReportFile == "" && flags.ReportFormat != "":
		flags.ReportFile = fmt.Sprintf("report-result.%s", flags.ReportFormat)
	case flags.ReportFile != "" && flags.ReportFormat == "":
		flags.ReportFormat = strings.ToLower(filepath.Ext(flags.ReportFile))[1:]
	}

	// Load configuration
	cfg, err := config.LoadConfig(flags.ConfigPath)
	if err != nil {
		mlog.Errorf("Error loading config from %s: %v", flags.ConfigPath, err)
		os.Exit(1)
	}
	mlog.Infof("Loaded configuration from %s", flags.ConfigPath)

	// Override with command line flags (only if provided)
	cfg.OverrideWithFlags(&flags.ConfigOverrideFlags)

	// Override random dataset config with command line flags
	// RandomEnable is handled separately because we need to distinguish between
	// "not set" and "set to false"
	if flags.RandomEnableSet {
		cfg.RandomDatasetVLLM.Enable = flags.RandomEnable
	}
	if flags.RandomInputLen > 0 {
		cfg.RandomDatasetVLLM.InputLength = flags.RandomInputLen
	}
	if flags.RandomOutputLen > 0 {
		cfg.RandomDatasetVLLM.OutputLength = flags.RandomOutputLen
	}

	// Validate configuration
	if cfg.Model.Provider == "" {
		mlog.Error("Provider must be specified in config file")
		os.Exit(1)
	}

	// Get system prompt
	systemPrompt := utils.GetSystemPrompt(&cfg.Model.SystemPromptTemplate)

	// Load or generate dataset based on configuration
	var dataset []provider.AnyParams

	if cfg.RandomDatasetVLLM.Enable {
		// Generate random dataset for vLLM
		dataset = generateRandomDataset(cfg, systemPrompt)
		mlog.Infof("Generated random dataset with input length %d tokens and output length %d tokens",
			cfg.RandomDatasetVLLM.InputLength, cfg.RandomDatasetVLLM.OutputLength)
	} else {
		// Load dataset from file
		dataset, err := utils.LoadDataset(cfg.Dataset.Path, cfg.Dataset.Type, systemPrompt)
		if err != nil {
			mlog.Errorf("Error loading dataset from %s: %v", cfg.Dataset.Path, err)
			os.Exit(1)
		}
		mlog.Infof("Loaded %d test cases from dataset %s", len(dataset), cfg.Dataset.Path)
	}

	// Create provider
	var prov provider.Provider
	switch cfg.Model.Provider {
	case "openai":
		prov = provider.NewOpenAIProvider(cfg.Model.ApiKey, cfg.Model.Endpoint, cfg.Model.Name, cfg.Test.Timeout)
	case "qwen":
		prov = provider.NewQwenProvider(cfg.Model.ApiKey, cfg.Model.Endpoint, cfg.Model.Name, cfg.Test.Timeout)
	default:
		mlog.Errorf("Unsupported provider: %s. Supported providers: openai, qwen", cfg.Model.Provider)
		os.Exit(1)
	}

	return &TestContext{
		Config:   cfg,
		Provider: prov,
		Dataset:  dataset,
	}
}

// generateRandomDataset generates a random dataset for vLLM testing
func generateRandomDataset(cfg *config.Config, systemPrompt string) []provider.AnyParams {
	// Generate a single test case with random prompt
	prompt, err := utils.GetRandomPromptByTokenCount(cfg.Model.Endpoint, cfg.RandomDatasetVLLM.InputLength)
	if err != nil {
		mlog.Warnf("Failed to generate exact token count: %v, using best effort", err)
	}

	messages := []interface{}{
		map[string]interface{}{
			"role":    "user",
			"content": prompt,
		},
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		messages = append([]interface{}{
			map[string]interface{}{
				"role":    "system",
				"content": systemPrompt,
			},
		}, messages...)
	}

	return []provider.AnyParams{
		{
			"messages":               messages,
			"max_tokens":             cfg.RandomDatasetVLLM.OutputLength,
			"ignore_eos":             true,
			"truncate_prompt_tokens": cfg.RandomDatasetVLLM.InputLength,
			"stream":                 true,
			"stream_options": map[string]interface{}{
				"include_usage": true,
			},
		},
	}
}
