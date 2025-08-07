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

	// Validate configuration
	if cfg.Model.Provider == "" {
		mlog.Error("Provider must be specified in config file")
		os.Exit(1)
	}

	// Get system prompt
	systemPrompt := utils.GetSystemPrompt(&cfg.Model.SystemPromptTemplate)

	// Load dataset with system prompt
	dataset, err := utils.LoadDataset(cfg.Dataset.Path, cfg.Dataset.Type, systemPrompt)
	if err != nil {
		mlog.Errorf("Error loading dataset from %s: %v", cfg.Dataset.Path, err)
		os.Exit(1)
	}
	mlog.Infof("Loaded %d test cases from dataset %s", len(dataset), cfg.Dataset.Path)

	// Create provider
	var prov provider.Provider
	switch cfg.Model.Provider {
	case "openai":
		prov = provider.NewOpenAIProvider(cfg.Model.ApiKey, cfg.Model.Endpoint, cfg.Test.Timeout)
		mlog.Infof("Created OpenAI provider with model %s", cfg.Model.Name)
	case "qwen":
		prov = provider.NewQwenProvider(cfg.Model.ApiKey, cfg.Model.Endpoint, cfg.Test.Timeout)
		mlog.Infof("Created Qwen provider with model %s", cfg.Model.Name)
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
