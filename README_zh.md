<!-- markdownlint-disable MD001 MD041 -->
<p align="center">
<img alt="gollmperf" src="./docs/assets/logos/logo1.png" width=50%>
</p>

# gollmperf - 专业LLM性能测试工具

[English](README.md) | 中文

## 项目概述

gollmperf是一个专业级的大语言模型(LLM)性能测试工具，旨在提供准确、全面的LLM性能评估。该工具支持多种LLM提供商，提供多维度性能指标，并具备企业级的测试能力。

## 核心功能特性

### 1. 多维度性能测试
- **吞吐量测试**: QPS (Queries Per Second) 测量
- **延迟测试**: TTFT (Time To First Token)、响应延迟、P50/P90/P99百分位数
- **质量测试**: 输出质量评估（可选）
- **稳定性测试**: 长时间运行稳定性验证

### 2. 多模型支持
- OpenAI (GPT系列)
- 阿里云 (通义千问系列)
- 自定义API端点

### 3. 高级测试模式
- **基础测试**: 标准性能测试
- **压力测试**: 逐步增加负载直到系统极限
- **性能测试**: 在多个并发级别下运行测试以找到最佳性能参数
- **稳定性测试**: 长时间持续运行测试
- **对比测试**: 多模型性能对比
- **场景测试**: 特定业务场景模拟

### 4. 专业统计指标
- **TTFT** (Time To First Token): 首字延迟
- **TPS** (Tokens Per Second): 每秒生成token数
- **成功率**: 请求成功率统计
- **错误分析**: 详细的错误类型和分布

### 5. 多样化报告输出
- 实时控制台输出
- JSON详细数据
- CSV表格数据
- HTML可视化报告

## 技术架构设计

### 核心模块

1. **测试引擎 (Engine)**
   - 负责执行各种测试任务
   - 精确控制并发请求数和负载
   - 预热阶段确保测试准确性

2. **数据收集器 (Collector)**
   - 收集性能数据和指标
   - 管理测试结果存储

3. **统计分析器 (Analyzer)**
   - 计算各种性能指标
   - 生成统计数据

4. **报告生成器 (Reporter)**
   - 生成多种格式的报告
   - 提供可视化展示

5. **配置管理器 (Config)**
   - 管理测试配置和参数
   - 支持YAML配置文件

6. **提供商接口 (Provider)**
   - 统一不同LLM提供商的接口
   - 支持扩展新的提供商

### 技术选型

- **编程语言**: Go (高性能、并发支持好)
- **并发模型**: goroutine + channel
- **HTTP客户端**: 标准库net/http
- **CLI框架**: Cobra
- **配置管理**: Viper + YAML
- **数据格式**: JSON、JSONL

## 项目结构

```
gollmperf/
├── cmd/                 # 命令行接口
├── configs/             # 配置文件示例
├── examples/            # 示例数据
├── internal/            # 核心模块
│   ├── engine/          # 测试引擎
│   ├── collector/       # 数据收集器
│   ├── analyzer/        # 统计分析器
│   ├── reporter/        # 报告生成器
│   ├── config/          # 配置管理
│   ├── provider/        # 提供商接口
│   └── utils/           # 工具函数
├── docs/                # 文档
└── main.go              # 主程序入口
```

## 使用方法

### 批量测试

```bash
# 使用--batch参数运行批量测试
./gollmperf run --config ./configs/example.yaml --batch
```

### 压力测试

```bash
# 运行压力测试
./gollmperf run --config ./configs/example.yaml 
```

### 性能测试

```bash
# 使用--perf参数运行性能测试模式
./gollmperf run --perf --config ./configs/example.yaml
```

在性能测试模式下，工具将在配置参数`perf_concurrency_group`中定义的多个并发级别下运行测试，以找到最佳性能参数。

### 命令行参数可以覆盖配置文件字段

`./gollmperf run -h`

```bash
  -k, --apikey string     API密钥
  -b, --batch             运行批量模式，执行数据集中的所有案例
  -d, --dataset string    数据集文件路径
  -e, --endpoint string   端点
  -f, --format string     报告格式 (json, csv, html) (默认为报告文件扩展名)
  -m, --model string      模型名称
  -p, --perf              运行性能模式，查找不同并发级别下的性能限制
  -P, --provider string   LLM提供商 (openai, qwen, 等) (默认 "openai")
  -r, --report string     报告文件路径 (输出报告到文件)
```

```bash
# 命令行参数覆盖配置文件字段
./gollmperf run --config ./configs/example.yaml --model gpt-3.5-turbo --dataset ./examples/test_cases.jsonl --report result.json --format json
```

### 对比测试

```bash
# 对比测试尚未实现，计划在后续版本中添加
# ./gollmperf compare --configs gpt35.yaml,gpt4.yaml,claude.yaml
```

### 使用示例

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

得到结果如下：

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
**默认当前目录得到一个 ./results/report.html 报告文件**

## 配置文件示例

```yaml
# Example configuration for gollmperf

# 测试配置
test:
  # 测试持续时间
  duration: 60s
  
  # 预热时间
  warmup: 10s
  
  # 并发请求数
  concurrency: 10
  
  # 请求超时时间
  timeout: 30s
  
  # 性能测试模式的并发级别组
  perf_concurrency_group: [1, 2, 4, 8, 16, 20, 32, 40, 48, 64]

# 模型配置
model:
  # 模型名称
  name: ${LLM_MODEL_NAME}
  
  # 提供商
  provider: openai
  
  # API端点 (可选，不指定则使用默认值)
  endpoint: ${LLM_API_ENDPOINT}

  # API密钥 (必需)
  api_key: ${LLM_API_KEY}

  # HTTP请求头，可添加任何额外的头部字段
  headers:
    Content-Type: application/json

  # HTTP请求参数模板，可添加任何额外的字段
  params_template:
    stream: true
    stream_options:
      include_usage: true
    extra_body:
      enable_thinking: false

# 数据集配置
dataset:
  # 数据集类型 (jsonl等)
  type: jsonl
  
  # 数据集文件路径
  path: ./examples/test_cases.jsonl

# 输出配置
output:
  # 输出格式 (json, csv, html)
  format: html
    
  # Output file path
  path: ./results/report.html
```

## 专业特性

### 1. 精确计时
- 微秒级精度计时
- 网络延迟分离测量
- GC影响排除

### 2. 负载控制
- 精确的并发控制
- QPS目标控制
- 自适应负载调节

### 3. 数据验证
- 响应内容验证
- Token计数准确性
- 错误分类统计

### 4. 企业级特性
- 多用户支持
- 权限控制
- 日志审计
- 集群部署支持

## 开发状态

项目已完成核心功能开发，包括：
- ✅ 项目结构和基础框架
- ✅ 配置管理模块
- ✅ 测试执行模块
- ✅ JSONL批量测试功能
- ✅ 统计分析模块
- ✅ 报告输出模块 (控制台、JSON、CSV、HTML)
- ✅ 带灵活参数配置的命令行接口
- ✅ 基本测试和验证
- ✅ OpenAI提供商实现
- ✅ 通义千问提供商实现
- ✅ 带错误分类的全面指标收集
- ✅ 批量测试模式
- ✅ 压力测试模式
- ✅ 性能测试模式

## 后续优化方向

1. **性能优化**: 进一步提升测试工具本身的性能
2. **功能扩展**: 添加更多测试模式和指标
3. **提供商支持**: 增加更多LLM提供商的支持
4. **可视化增强**: 提供更丰富的图表和仪表板
5. **企业功能**: 添加用户管理、权限控制等企业级功能
6. **压力测试实现**: 完成压力测试模式
7. **对比测试实现**: 添加多模型性能对比功能

## 总结

gollmperf为LLM性能测试提供了一个专业、准确、易用的解决方案。通过模块化设计和清晰的架构，该工具不仅满足了当前的测试需求，还具备了良好的扩展性，可以适应未来更多的测试场景和需求。

## 许可证

本项目采用Apache License 2.0许可证 - 详情请查看[LICENSE](LICENSE)文件。
