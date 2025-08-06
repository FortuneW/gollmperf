<!-- markdownlint-disable MD001 MD041 -->
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./docs/assets/logos/png.jpg">
    <img alt="gollmperf" src="./docs/assets/logos/logo1.png" width=55%>
  </picture>
</p>

# gollmperf - Professional LLM Performance Testing Tool

English | [中文](README_zh.md)

## Project Overview

gollmperf is a professional-grade Large Language Model (LLM) performance testing tool designed to provide accurate and comprehensive LLM performance evaluation. The tool supports multiple LLM providers, offers multi-dimensional performance metrics, and features enterprise-level testing capabilities.

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
gollmperf/
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
./gollmperf run --config ./configs/example.yaml
```

### Stress Testing

```bash
# Run stress test mode with --stress flag
./gollmperf run --stress --config ./configs/example.yaml
```

### Command args can override config file fields

`./gollmperf run -h`

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
./gollmperf run --config ./configs/example.yaml --model gpt-3.5-turbo --dataset ./examples/test_cases.jsonl --report result.json --format json
```

### Comparative Testing

```bash
# Comparative testing is not yet implemented but planned for future releases
# ./gollmperf compare --configs gpt35.yaml,gpt4.yaml,claude.yaml
```


### Usage Examples

```bash
git clone https://github.com/FortuneW/gollmperf.git

cd gollmperf
go mod tidy
go build

export LLM_API_ENDPOINT="https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
export LLM_API_KEY="sk-xxx"
export LLM_MODEL_NAME="qwen-plus"

./gollmperf run -c configs/example.yaml
```

The results are as follows:

``` bash
[INF] 2025-08-06T11:13:15.830Z [reporter] ========== gollmperf Performance Report ==========
[INF] 2025-08-06T11:13:15.830Z [reporter] Total Duration: 4.212978738s
[INF] 2025-08-06T11:13:15.830Z [reporter] Total Requests: 5
[INF] 2025-08-06T11:13:15.830Z [reporter] Successful Requests: 5
[INF] 2025-08-06T11:13:15.830Z [reporter] Failed Requests: 0
[INF] 2025-08-06T11:13:15.830Z [reporter] Success Rate: 100.00%
[INF] 2025-08-06T11:13:15.830Z [reporter] QPS: 1.19
[INF] 2025-08-06T11:13:15.830Z [reporter] Tokens per second: 197.01
[INF] 2025-08-06T11:13:15.830Z [reporter] Average Latency: 3.650148772s
[INF] 2025-08-06T11:13:15.830Z [reporter] Latency P50: 3.822069212s
[INF] 2025-08-06T11:13:15.830Z [reporter] Latency P90: 4.212863638s
[INF] 2025-08-06T11:13:15.830Z [reporter] Latency P99: 4.212863638s
[INF] 2025-08-06T11:13:15.830Z [reporter] Average Request Tokens: 22.00
[INF] 2025-08-06T11:13:15.830Z [reporter] Average Response Tokens: 144.00
[INF] 2025-08-06T11:13:15.830Z [reporter] Average First Token Latency: 488.073612ms
[INF] 2025-08-06T11:13:15.830Z [reporter] First Token Latency P50: 512.180236ms
[INF] 2025-08-06T11:13:15.830Z [reporter] First Token Latency P90: 583.377086ms
[INF] 2025-08-06T11:13:15.830Z [reporter] First Token Latency P99: 583.377086ms
```
**By default, a ./results/report.html report file will be generated in the current directory**

## Configuration Example

```yaml
# Example configuration for gollmperf

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
    
  # Output file path
  path: ./results/report.html
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

gollmperf provides a professional, accurate, and user-friendly solution for LLM performance testing. Through modular design and a clear architecture, the tool not only meets current testing requirements but also has good extensibility to adapt to future testing scenarios and needs.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
