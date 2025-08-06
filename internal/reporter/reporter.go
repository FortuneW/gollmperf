package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/FortuneW/gollmperf/internal/analyzer"
	"github.com/FortuneW/qlog"
)

var mlog = qlog.GetRLog("reporter")

// Reporter generates reports from analysis results
type Reporter struct {
	metrics *analyzer.Metrics
}

// NewReporter creates a new reporter
func NewReporter(metrics *analyzer.Metrics) *Reporter {
	return &Reporter{
		metrics: metrics,
	}
}

// GenerateConsoleReport generates a console report
func (r *Reporter) GenerateConsoleReport() {
	mlog.Info("========== gollmperf Performance Report ==========")
	mlog.Infof("Total Duration: %v", r.metrics.TotalDuration)
	mlog.Infof("Total Requests: %d", r.metrics.TotalRequests)
	mlog.Infof("Successful Requests: %d", r.metrics.SuccessfulRequests)
	mlog.Infof("Failed Requests: %d", r.metrics.FailedRequests)
	mlog.Infof("Success Rate: %.2f%%", r.metrics.SuccessRate)

	if r.metrics.SuccessfulRequests > 0 {
		mlog.Infof("QPS: %.2f", r.metrics.QPS)
		mlog.Infof("Tokens per second: %.2f", r.metrics.TokensPerSecond)
		mlog.Infof("Average Latency: %v", r.metrics.AverageLatency)
		mlog.Infof("Latency P50: %v", r.metrics.LatencyP50)
		mlog.Infof("Latency P90: %v", r.metrics.LatencyP90)
		mlog.Infof("Latency P99: %v", r.metrics.LatencyP99)

		mlog.Infof("Average Request Tokens: %.2f", r.metrics.AverageRequestTokens)
		mlog.Infof("Average Response Tokens: %.2f", r.metrics.AverageResponseTokens)

		if r.metrics.AverageFirstTokenLatency > 0 {
			mlog.Infof("Average First Token Latency: %v", r.metrics.AverageFirstTokenLatency)
			mlog.Infof("First Token Latency P50: %v", r.metrics.FirstTokenLatencyP50)
			mlog.Infof("First Token Latency P90: %v", r.metrics.FirstTokenLatencyP90)
			mlog.Infof("First Token Latency P99: %v", r.metrics.FirstTokenLatencyP99)
		}
	}

	if len(r.metrics.ErrorTypeCounts) > 0 {
		mlog.Info("Error Type Distribution:")
		for error, count := range r.metrics.ErrorTypeCounts {
			mlog.Errorf("  %s: %d", error, count)
		}
	}

	mlog.Info("========== End of Report ==========")
}

// GenerateJSONReport generates a JSON report
func (r *Reporter) GenerateJSONReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(r.metrics); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// GenerateCSVReport generates a CSV report
func (r *Reporter) GenerateCSVReport(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write CSV header
	header := "total_requests,successful_requests,failed_requests,success_rate,qps,tokens_per_second," +
		"average_latency,latency_p50,latency_p90,latency_p99," +
		"average_request_tokens,average_response_tokens," +
		"average_first_token_latency,first_token_latency_p50,first_token_latency_p90,first_token_latency_p99\n"

	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data row
	row := fmt.Sprintf("%d,%d,%d,%.2f,%.2f,%.2f,%d,%d,%d,%d,%.2f,%.2f,%d,%d,%d,%d\n",
		r.metrics.TotalRequests,
		r.metrics.SuccessfulRequests,
		r.metrics.FailedRequests,
		r.metrics.SuccessRate,
		r.metrics.QPS,
		r.metrics.TokensPerSecond,
		r.metrics.AverageLatency.Milliseconds(),
		r.metrics.LatencyP50.Milliseconds(),
		r.metrics.LatencyP90.Milliseconds(),
		r.metrics.LatencyP99.Milliseconds(),
		r.metrics.AverageRequestTokens,
		r.metrics.AverageResponseTokens,
		r.metrics.AverageFirstTokenLatency.Milliseconds(),
		r.metrics.FirstTokenLatencyP50.Milliseconds(),
		r.metrics.FirstTokenLatencyP90.Milliseconds(),
		r.metrics.FirstTokenLatencyP99.Milliseconds(),
	)

	if _, err := file.WriteString(row); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// GenerateHTMLReport generates an HTML report
func (r *Reporter) GenerateHTMLReport(filename string) error {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>gollmperf Performance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1, h2 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; }
        .metric-value { font-weight: bold; }
    </style>
</head>
<body>
    <h1>gollmperf Performance Report</h1>
    
    <h2>Summary</h2>
    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Total Duration</td><td class="metric-value">{{.TotalDuration.Milliseconds}}</td></tr>
        <tr><td>Total Requests</td><td class="metric-value">{{.TotalRequests}}</td></tr>
        <tr><td>Successful Requests</td><td class="metric-value">{{.SuccessfulRequests}}</td></tr>
        <tr><td>Failed Requests</td><td class="metric-value">{{.FailedRequests}}</td></tr>
        <tr><td>Success Rate</td><td class="metric-value">{{printf "%.2f" .SuccessRate}}%</td></tr>
    </table>
    
    {{if gt .SuccessfulRequests 0}}
    <h2>Performance Metrics</h2>
    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>QPS</td><td class="metric-value">{{printf "%.2f" .QPS}}</td></tr>
        <tr><td>Tokens per second</td><td class="metric-value">{{printf "%.2f" .TokensPerSecond}}</td></tr>
        <tr><td>Average Latency</td><td class="metric-value">{{.AverageLatency.Milliseconds}}</td></tr>
        <tr><td>Latency P50</td><td class="metric-value">{{.LatencyP50.Milliseconds}}</td></tr>
        <tr><td>Latency P90</td><td class="metric-value">{{.LatencyP90.Milliseconds}}</td></tr>
        <tr><td>Latency P99</td><td class="metric-value">{{.LatencyP99.Milliseconds}}</td></tr>
    </table>
    
    <h2>Token Metrics</h2>
    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Average Request Tokens</td><td class="metric-value">{{printf "%.2f" .AverageRequestTokens}}</td></tr>
        <tr><td>Average Response Tokens</td><td class="metric-value">{{printf "%.2f" .AverageResponseTokens}}</td></tr>
    </table>
    
    {{if gt .AverageFirstTokenLatency 0}}
    <h2>Streaming Metrics</h2>
    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Average First Token Latency</td><td class="metric-value">{{.AverageFirstTokenLatency.Milliseconds}}</td></tr>
        <tr><td>First Token Latency P50</td><td class="metric-value">{{.FirstTokenLatencyP50.Milliseconds}}</td></tr>
        <tr><td>First Token Latency P90</td><td class="metric-value">{{.FirstTokenLatencyP90.Milliseconds}}</td></tr>
        <tr><td>First Token Latency P99</td><td class="metric-value">{{.FirstTokenLatencyP99.Milliseconds}}</td></tr>
    </table>
    {{end}}
    {{end}}
    
    {{if .ErrorTypeCounts}}
    <h2>Error Distribution</h2>
    <table>
        <tr><th>Error</th><th>Count</th></tr>
        {{range $error, $count := .ErrorTypeCounts}}
        <tr><td>{{$error}}</td><td>{{$count}}</td></tr>
        {{end}}
    </table>
    {{end}}
</body>
</html>`

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, r.metrics); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// GenerateFileReport generates a report in the specified format
func (r *Reporter) GenerateFileReport(reportFile, reportFormat string) error {
	_ = os.MkdirAll(filepath.Dir(reportFile), 0755)
	if !strings.HasSuffix(strings.ToLower(reportFile), reportFormat) {
		reportFile = reportFile + "." + reportFormat
	}
	switch reportFormat {
	case "json":
		if err := r.GenerateJSONReport(reportFile); err != nil {
			return fmt.Errorf("failed to generate JSON report: %w", err)
		}
		mlog.Infof("JSON report generated: %s", reportFile)
	case "csv":
		if err := r.GenerateCSVReport(reportFile); err != nil {
			return fmt.Errorf("failed to generate CSV report: %w", err)
		}
		mlog.Infof("CSV report generated: %s", reportFile)
	case "html":
		if err := r.GenerateHTMLReport(reportFile); err != nil {
			return fmt.Errorf("failed to generate HTML report: %w", err)
		}
		mlog.Infof("HTML report generated: %s", reportFile)
	default:
		return fmt.Errorf("unsupported report format: %s. Supported formats: json, csv, html", reportFormat)
	}
	return nil
}
