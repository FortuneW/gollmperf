// Translation dictionary
const translations = {
    en: {
        "concurrentTestComparison": "Concurrent Test Comparison",
        "bestPerformance": "Best Performance",
        "highestQPS": "Highest QPS",
        "noDataAvailable": "No data available",
        "bestThroughput": "Best Throughput",
        "highestTokensPerSecond": "Highest Tokens per Second",
        "e2eLatencyBottleneck": "E2E Latency Bottleneck",
        "bottleneckDetected": "Bottleneck Detected",
        "recommended": "Recommended",
        "optimalConcurrency": "Optimal Concurrency",
        "detailedComparison": "Detailed Comparison",
        "concurrency": "Concurrency",
        "requests": "Requests",
        "duration": "Duration",
        "qps": "QPS",
        "tokensPerSec": "Tokens/sec",
        "e2eLatency": "E2E Latency",
        "firstTokenLatency": "First Token Latency",
        "tokenMetrics": "Token Metrics",
        "average": "Average",
        "p50": "P50",
        "p90": "P90",
        "p99": "P99",
        "request": "Request",
        "response": "Response",
        "performanceMetricsChart": "Performance Metrics Chart",
        "latencyDistributionChart": "Latency Distribution Chart",
        "firstTokenLatencyChart": "First Token Latency Chart",
        "errorStatistics": "Error Statistics",
        "errorRate": "Error Rate",
        "errorTypeDistribution": "Error Type Distribution"
    },
    zh: {
        "concurrentTestComparison": "并发测试比较",
        "bestPerformance": "最佳性能",
        "highestQPS": "最高 QPS",
        "noDataAvailable": "无可用数据",
        "bestThroughput": "最佳吞吐量",
        "highestTokensPerSecond": "最高每秒令牌数",
        "e2eLatencyBottleneck": "端到端延迟瓶颈",
        "bottleneckDetected": "检测到瓶颈",
        "recommended": "推荐",
        "optimalConcurrency": "最优并发数",
        "detailedComparison": "详细比较",
        "concurrency": "并发数",
        "requests": "请求数",
        "duration": "持续时间",
        "qps": "QPS",
        "tokensPerSec": "Tokens/秒",
        "e2eLatency": "端到端延迟",
        "firstTokenLatency": "首Token延迟",
        "tokenMetrics": "Token指标",
        "average": "平均",
        "p50": "P50",
        "p90": "P90",
        "p99": "P99",
        "request": "请求",
        "response": "响应",
        "performanceMetricsChart": "性能指标图表",
        "latencyDistributionChart": "延迟分布图表",
        "firstTokenLatencyChart": "首Token延迟图表",
        "errorStatistics": "错误统计",
        "errorRate": "错误率",
        "errorTypeDistribution": "错误类型分布"
    }
};

// Language switching function
function switchLanguage(lang) {
    // Update active button
    document.getElementById('lang-en').classList.toggle('active', lang === 'en');
    document.getElementById('lang-zh').classList.toggle('active', lang === 'zh');

    // Update all elements with data-i18n attribute
    const elements = document.querySelectorAll('[data-i18n]');
    elements.forEach(element => {
        const key = element.getAttribute('data-i18n');
        if (translations[lang] && translations[lang][key]) {
            if (element.tagName === 'H1' || element.tagName === 'H2' || element.tagName === 'H3') {
                element.textContent = translations[lang][key];
            } else {
                element.innerHTML = translations[lang][key];
            }
        }
    });

    // Update the report title specifically
    document.getElementById('report-title').textContent = lang === 'en' ?
        'goLLMPerf Performance Report' : 'goLLMPerf 性能报告';
}

// Add event listeners to language buttons
document.getElementById('lang-en').addEventListener('click', () => switchLanguage('en'));
document.getElementById('lang-zh').addEventListener('click', () => switchLanguage('zh'));

// Set initial language to English
switchLanguage('en');