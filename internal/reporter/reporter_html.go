package reporter

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

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
	reportTmplJS, _ := templateFS.ReadFile("templates/js/report.tmpl.js")

	// Create wrapper data for template
	data := &ReporterData{
		ReporterData:  r.concurrentComparison,
		ChartTmplCSS:  string(chartTmplCSS),
		ReportTmplCSS: string(reportTmplCSS),
		ChartTmplJS:   string(jsFileContent),
		ReportTmplJS:  string(reportTmplJS),
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
