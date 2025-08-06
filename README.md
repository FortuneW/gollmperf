# LLMPerf - Professional LLM Performance Testing Tool

English | [中文](README_zh.md)

## Project Overview

LLMPerf is a professional-grade Large Language Model (LLM) performance testing tool designed to provide accurate and comprehensive LLM performance evaluation. The tool supports multiple LLM providers, offers multi-dimensional performance metrics, and features enterprise-level testing capabilities.

## Core Features

### 1. Multi-dimensional Performance Testing
- **Throughput Testing**: QPS (Queries Per Second) measurement
- **Latency Testing**: TTFT (Time To First Token), response latency, P50/P90/P99 percentiles
- **Quality Testing**: Output quality assessment (optional)
- **Stability Testing**: Long-term runtime stability verification

### 2. Multi-model Support
- OpenAI (GPT series)
- Qwen (Tongyi Qianwen series)
- Custom API endpoints

### 3. Advanced Testing Modes
- **Basic Testing**: Standard performance testing
- **Stress Testing**: Gradually increase load until system limits
- **Stability Testing**: Long-term continuous runtime testing
- **Comparative Testing**: Multi-model performance comparison
- **Scenario Testing**: Specific business scenario simulation

### 4. Professional Metrics
- **TTFT** (Time To First Token): First token latency
- **TPS** (Tokens Per Second): Tokens generated per second
- **Success Rate**: Request success rate statistics
- **Error Analysis**: Detailed error type and distribution

### 5. Diverse Report Output
- Real-time console output
- JSON detailed data
- CSV tabular data
- HTML visualization reports

## Technical Architecture Design

### Core Modules

1. **Test Engine**
   - Executes various test tasks
   - Precisely controls concurrent requests and load
   - Warm-up phase ensures test accuracy

2. **Data Collector**
   - Collects performance data and metrics
   - Manages test result storage

3. **Statistical Analyzer**
   - Calculates various performance metrics
   - Generates statistical data

4. **Report Generator**
   - Generates reports in multiple formats
   - Provides visualization display

5. **Configuration Manager**
   - Manages test configurations and parameters
   - Supports YAML configuration files

6. **Provider Interface**
   - Unifies interfaces for different LLM providers
   - Supports extension for new providers

### Technology Stack

- **Programming Language**: Go (High performance, excellent concurrency support)
- **Concurrency Model**: goroutine + channel
- **HTTP Client**: Standard library net/http
- **CLI Framework**: Cobra
- **Configuration Management**: Viper + YAML
- **Data Format**: JSON, JSONL

## Project Structure

```
llmperf/
├── cmd/                 # Command-line interface
├── configs/             # Configuration file examples
├── examples/            # Sample data
├── internal/            # Core modules
│   ├── engine/          # Test engine
│   ├── collector/       # Data collector
│   ├── analyzer/        # Statistical analyzer
│   ├── reporter/        # Report generator
│   ├── config/          # Configuration management
│   ├── provider/        # Provider interface
│   └── utils/           # Utility functions
├── docs/                # Documentation
└── main.go              # Main program entry
```

## Usage

### Batch Testing

```bash
# Using configuration file for batch mode
./llmperf run --config ./configs/example.yaml
```

### Stress Testing

```bash
# Run stress test mode with --stress flag
./llmperf run --stress --config ./configs/example.yaml
```

### Command args can override config file fields

`./llmperf run -h`

```bash
  -k, --apikey string     API key
  -d, --dataset string    Dataset file path
  -e, --endpoint string   Endpoint
  -f, --format string     Report format (json, csv, html) (default as report file extension)
  -m, --model string      Model name
  -r, --report string     Report file path (output report to file)
```

```bash
# Command args override config file fields
./llmperf run --config ./configs/example.yaml --model gpt-3.5-turbo --dataset ./examples/test_cases.jsonl --report result.json --format json
```

### Comparative Testing

```bash
# Comparative testing is not yet implemented but planned for future releases
# ./llmperf compare --configs gpt35.yaml,gpt4.yaml,claude.yaml
```

## Configuration Example

```yaml
# Example configuration for LLMPerf

# Test configuration
test:
  # Duration of the test
  duration: 60s
  
  # Warmup period before starting measurements
  warmup: 10s
  
  # Number of concurrent requests
  concurrency: 10
  
  # Timeout for each request
  timeout: 30s

# Model configuration
model:
  # Model name
  name: ${LLM_MODEL_NAME}
  
  # Provider (openai, qwen, etc.)
  provider: openai
  
  # API endpoint (optional, uses default if not specified)
  endpoint: ${LLM_API_ENDPOINT}

  # API key (required)
  api_key: ${LLM_API_KEY}

  # http headers, with any additional header fields
  headers:
    Content-Type: application/json

  # http request params template, with any additional fields
  params_template:
    stream: true
    stream_options:
      include_usage: true
    extra_body:
      enable_thinking: false

# Dataset configuration
dataset:
  # Type of dataset (jsonl, etc.)
  type: jsonl
  
  # Path to dataset file
  path: ./examples/test_cases.jsonl

# Output configuration
output:
  # Output formats (json, csv, html)
  format: html
    
  # Output directory
  path: ./results
```

## Professional Features

### 1. Precise Timing
- Microsecond-level timing accuracy
- Network latency separation measurement
- GC impact exclusion

### 2. Load Control
- Precise concurrency control
- QPS target control
- Adaptive load regulation

### 3. Data Validation
- Response content validation
- Token count accuracy
- Error classification statistics

### 4. Enterprise Features
- Multi-user support
- Permission control
- Log auditing
- Cluster deployment support

## Development Status

The project has completed core functionality development, including:
- ✅ Project structure and basic framework
- ✅ Configuration management module
- ✅ Test execution module
- ✅ JSONL batch testing functionality
- ✅ Statistical analysis module
- ✅ Report output module (console, JSON, CSV, HTML)
- ✅ Command-line interface with flexible parameter configuration
- ✅ Basic testing and validation
- ✅ OpenAI provider implementation
- ✅ Qwen provider implementation
- ✅ Comprehensive metrics collection with error categorization
- ✅ Batch testing mode
- ✅ Stress testing mode
- ✅ Performance testing mode (planned implementation)

## Future Optimization Directions

1. **Performance Optimization**: Further enhance the testing tool's own performance
2. **Feature Expansion**: Add more testing modes and metrics
3. **Provider Support**: Increase support for more LLM providers
4. **Visualization Enhancement**: Provide richer charts and dashboards
5. **Enterprise Features**: Add user management, permission control, and other enterprise-level features
6. **Stress Testing Implementation**: Complete the stress testing mode
7. **Comparative Testing Implementation**: Add multi-model performance comparison capabilities

## Summary

LLMPerf provides a professional, accurate, and user-friendly solution for LLM performance testing. Through modular design and a clear architecture, the tool not only meets current testing requirements but also has good extensibility to adapt to future testing scenarios and needs.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
