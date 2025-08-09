package reporter

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"bytes"

	"github.com/FortuneW/gollmperf/internal/analyzer"
	"github.com/FortuneW/qlog"
)

var mlog = qlog.GetRLog("reporter")

// ReporterData is a wrapper struct for template data
type ReporterData struct {
	ReporterData  *ConcurrentComparison `json:"reporter_data,omitempty"`
	ReportTmplCSS string                `json:"-"`
	ChartTmplCSS  string                `json:"-"`
	ChartTmplJS   string                `json:"-"`
}

// Reporter generates reports from analysis results
type Reporter struct {
	metrics              *analyzer.Metrics
	concurrentComparison *ConcurrentComparison
}

// NewReporter creates a new reporter
func NewReporter() *Reporter {
	return &Reporter{
		concurrentComparison: &ConcurrentComparison{},
	}
}

// AddNewMetrics adds new metrics to the reporter
func (r *Reporter) AddNewMetrics(concurrency int, metrics *analyzer.Metrics) {
	r.concurrentComparison.TestResults = append(r.concurrentComparison.TestResults, ConcurrentTestResult{
		Concurrency: concurrency,
		Metrics:     metrics,
	})
	r.metrics = metrics // current metrics
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

	if err := encoder.Encode(r.concurrentComparison); err != nil {
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
	header := "concurrency,total_requests,successful_requests,failed_requests,success_rate,qps,tokens_per_second," +
		"average_latency,latency_p50,latency_p90,latency_p99," +
		"average_request_tokens,average_response_tokens," +
		"average_first_token_latency,first_token_latency_p50,first_token_latency_p90,first_token_latency_p99\n"

	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, result := range r.concurrentComparison.TestResults {
		// Write data row
		row := fmt.Sprintf("%d,%d,%d,%d,%.2f,%.2f,%.2f,%d,%d,%d,%d,%.2f,%.2f,%d,%d,%d,%d\n",
			result.Concurrency,
			result.Metrics.TotalRequests,
			result.Metrics.SuccessfulRequests,
			result.Metrics.FailedRequests,
			result.Metrics.SuccessRate,
			result.Metrics.QPS,
			result.Metrics.TokensPerSecond,
			result.Metrics.AverageLatency.Milliseconds(),
			result.Metrics.LatencyP50.Milliseconds(),
			result.Metrics.LatencyP90.Milliseconds(),
			result.Metrics.LatencyP99.Milliseconds(),
			result.Metrics.AverageRequestTokens,
			result.Metrics.AverageResponseTokens,
			result.Metrics.AverageFirstTokenLatency.Milliseconds(),
			result.Metrics.FirstTokenLatencyP50.Milliseconds(),
			result.Metrics.FirstTokenLatencyP90.Milliseconds(),
			result.Metrics.FirstTokenLatencyP99.Milliseconds(),
		)

		if _, err := file.WriteString(row); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}
	return nil
}

//go:embed templates/*
var templateFS embed.FS

// GenerateHTMLReport generates an HTML report
func (r *Reporter) GenerateHTMLReport(filename string) error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(filename)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Read the template file from embedded filesystem
	templateData, err := templateFS.ReadFile("templates/report.tmpl.html")
	if err != nil {
		return fmt.Errorf("failed to read template file from embedded filesystem: %w", err)
	}

	tmpl, err := template.New("report").Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	jsFileContent, _ := templateFS.ReadFile("templates/js/chart.tmpl.js")
	chartTmplCSS, _ := templateFS.ReadFile("templates/css/chart.tmpl.css")
	reportTmplCSS, _ := templateFS.ReadFile("templates/css/report.tmpl.css")

	// Create wrapper data for template
	data := &ReporterData{
		ReporterData:  r.concurrentComparison,
		ChartTmplCSS:  string(chartTmplCSS),
		ReportTmplCSS: string(reportTmplCSS),
		ChartTmplJS:   string(jsFileContent),
	}
	
	// Process JavaScript template with Go template engine
	jsTmpl, err := template.New("chart.js").Parse(data.ChartTmplJS)
	if err != nil {
		return fmt.Errorf("failed to parse JavaScript template: %w", err)
	}
	
	// Execute JavaScript template
	var jsBuffer bytes.Buffer
	if err := jsTmpl.Execute(&jsBuffer, data); err != nil {
		return fmt.Errorf("failed to execute JavaScript template: %w", err)
	}
	
	// Update the ChartTmplJS field with processed JavaScript
	data.ChartTmplJS = jsBuffer.String()

	// Execute template with wrapper data
	if err := tmpl.Execute(file, data); err != nil {
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
