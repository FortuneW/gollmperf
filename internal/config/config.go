package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/FortuneW/qlog"
	"github.com/spf13/viper"
)

// GenerateDefaultConfig generates a default configuration file
func GenerateDefaultConfig(filePath string) error {
	// Create a new config with default values
	config := NewConfig()

	// Add default values for test config
	config.Test.Duration = 60 * time.Second
	config.Test.Warmup = 10 * time.Second
	config.Test.Concurrency = 10
	config.Test.RequestsPerConcurrency = 100
	config.Test.Timeout = 30 * time.Second
	config.Test.PerfConcurrencyGroup = []int{1, 2, 4, 8, 16, 20, 32, 40, 48, 64}

	// Add default values for model config
	config.Model.Name = "${LLM_MODEL_NAME}"
	config.Model.Provider = "openai"
	config.Model.Endpoint = "${LLM_API_ENDPOINT}"
	config.Model.ApiKey = "${LLM_API_KEY}"
	config.Model.Headers = map[string]string{
		"Content-Type": "application/json",
	}

	// Add default values for system_prompt_template
	config.Model.SystemPromptTemplate = SystemPromptTemplate{
		Enable:  false,
		Content: "You are a helpful assistant.",
		Path:    "./examples/system_prompt.md",
	}

	// Add default values for model params
	config.Model.ParamsTemplate = map[string]interface{}{
		"stream": true,
		"stream_options": map[string]interface{}{
			"include_usage": true,
		},
		"extra_body": map[string]interface{}{
			"enable_thinking": false,
		},
	}

	// Add default values for dataset
	config.Dataset.Path = "./examples/test_cases.jsonl"

	// Add default values for output
	config.Output.Path = "./results/report.html"
	config.Output.BatchResultPath = "./results/batch_results.jsonl"

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Previously existing code

var mlog = qlog.GetRLog("config")

// Config represents the complete configuration for LLMPerf
type Config struct {
	Test    TestConfig    `yaml:"test"`
	Model   ModelConfig   `yaml:"model"`
	Dataset DatasetConfig `yaml:"dataset"`
	Output  OutputConfig  `yaml:"output"`
}

// TestConfig represents test configuration
type TestConfig struct {
	Duration               time.Duration
	Warmup                 time.Duration
	Concurrency            int
	RequestsPerConcurrency int `mapstructure:"requests_per_concurrency"`
	Timeout                time.Duration
	PerfConcurrencyGroup   []int `mapstructure:"perf_concurrency_group"`
}

// SystemPromptTemplate represents the system prompt configuration
// It supports either direct content or a file path
type SystemPromptTemplate struct {
	Enable  bool   `yaml:"enable"`
	Content string `yaml:"content"`
	Path    string `yaml:"path"`
}

// ModelConfig represents model configuration
type ModelConfig struct {
	Name                 string
	Provider             string
	Endpoint             string
	Headers              map[string]string
	ApiKey               string                 `mapstructure:"api_key"`
	ParamsTemplate       map[string]interface{} `mapstructure:"params_template"`
	SystemPromptTemplate SystemPromptTemplate   `mapstructure:"system_prompt_template"`
}

// DatasetConfig represents dataset configuration
type DatasetConfig struct {
	Type string
	Path string
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Format          string
	Path            string
	BatchResultPath string `mapstructure:"batch_result_path"`
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Create a new viper instance
	v := viper.New()

	// Set config file
	v.SetConfigFile(configPath)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create new config instance
	config := NewConfig()

	// Unmarshal config
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Handle environment variable substitution for model.name
	if modelName := v.GetString("model.name"); modelName != "" {
		if len(modelName) > 3 && modelName[:2] == "${" && modelName[len(modelName)-1:] == "}" {
			envVar := modelName[2 : len(modelName)-1]
			config.Model.Name = os.Getenv(envVar)
		}
	}

	// Handle environment variable substitution for model.api_key
	if apiKey := v.GetString("model.api_key"); apiKey != "" {
		if len(apiKey) > 3 && apiKey[:2] == "${" && apiKey[len(apiKey)-1:] == "}" {
			envVar := apiKey[2 : len(apiKey)-1]
			config.Model.ApiKey = os.Getenv(envVar)
		}
	}

	// Handle environment variable substitution for model.endpoint
	if endpoint := v.GetString("model.endpoint"); endpoint != "" {
		if len(endpoint) > 3 && endpoint[:2] == "${" && endpoint[len(endpoint)-1:] == "}" {
			envVar := endpoint[2 : len(endpoint)-1]
			config.Model.Endpoint = os.Getenv(envVar)
		}
	}

	if config.Model.SystemPromptTemplate.Enable {
		if config.Model.SystemPromptTemplate.Content != "" && config.Model.SystemPromptTemplate.Path != "" {
			mlog.Warnf("Both content and path are set for system_prompt_template, content will take precedence")
		}
	}

	return config, nil
}

// OverrideWithFlags overrides config values with command line flags
func (c *Config) OverrideWithFlags(flags *ConfigOverrideFlags) {
	if flags.Provider != "" {
		c.Model.Provider = flags.Provider
	}
	if flags.Model != "" {
		c.Model.Name = flags.Model
	}
	if flags.Dataset != "" {
		c.Dataset.Path = flags.Dataset
	}
	if flags.ApiKey != "" {
		c.Model.ApiKey = flags.ApiKey
	}
	if flags.Endpoint != "" {
		c.Model.Endpoint = flags.Endpoint
	}
	if flags.ReportFile != "" {
		c.Output.Path = flags.ReportFile
	}
	if flags.ReportFormat != "" {
		c.Output.Format = flags.ReportFormat
	}
	if flags.BatchResultFile != "" {
		c.Output.BatchResultPath = flags.BatchResultFile
	}
}

// ConfigOverrideFlags holds the command line flags for overriding config values
type ConfigOverrideFlags struct {
	Provider        string
	Model           string
	Dataset         string
	ApiKey          string
	Endpoint        string
	ReportFile      string
	ReportFormat    string
	BatchResultFile string
}
