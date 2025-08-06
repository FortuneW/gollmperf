package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete configuration for LLMPerf
type Config struct {
	Test    TestConfig    `yaml:"test"`
	Model   ModelConfig   `yaml:"model"`
	Dataset DatasetConfig `yaml:"dataset"`
	Output  OutputConfig  `yaml:"output"`
}

// TestConfig represents test configuration
type TestConfig struct {
	Duration    time.Duration `yaml:"duration"`
	Warmup      time.Duration `yaml:"warmup"`
	Concurrency int           `yaml:"concurrency"`
	Timeout     time.Duration `yaml:"timeout"`
}

// ModelConfig represents model configuration
type ModelConfig struct {
	Name           string                 `yaml:"name"`
	Provider       string                 `yaml:"provider"`
	Endpoint       string                 `yaml:"endpoint"`
	ApiKey         string                 `yaml:"api_key"`
	Headers        map[string]string      `yaml:"headers"`
	ParamsTemplate map[string]interface{} `yaml:"-"`
}

// DatasetConfig represents dataset configuration
type DatasetConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Format string `yaml:"format"`
	Path   string `yaml:"path"`
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		Test: TestConfig{
			Concurrency: 1,
		},
		Model: ModelConfig{},
		Dataset: DatasetConfig{
			Type: "jsonl",
		},
		Output: OutputConfig{
			Format: "html",
			Path:   "./results",
		},
	}
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

	// Handle environment variable substitution for model.endpoint
	if endpoint := v.GetString("model.endpoint"); endpoint != "" {
		if len(endpoint) > 3 && endpoint[:2] == "${" && endpoint[len(endpoint)-1:] == "}" {
			envVar := endpoint[2 : len(endpoint)-1]
			config.Model.Endpoint = os.Getenv(envVar)
		}
	}

	// Handle environment variable substitution for model.api_key
	if apiKey := v.GetString("model.api_key"); apiKey != "" {
		if len(apiKey) > 3 && apiKey[:2] == "${" && apiKey[len(apiKey)-1:] == "}" {
			envVar := apiKey[2 : len(apiKey)-1]
			config.Model.ApiKey = os.Getenv(envVar)
		}
	}

	// Handle ParamsTemplate
	if paramsTemplate := v.GetStringMap("model.params_template"); paramsTemplate != nil {
		config.Model.ParamsTemplate = paramsTemplate
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
}

// ConfigOverrideFlags holds the command line flags for overriding config values
type ConfigOverrideFlags struct {
	Provider     string
	Model        string
	Dataset      string
	ApiKey       string
	Endpoint     string
	ReportFile   string
	ReportFormat string
}
