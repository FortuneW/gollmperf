package reporter

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (r *Reporter) GenerateConsoleTableReport() {
	// Check if there are test results
	if len(r.concurrentComparison.TestResults) == 0 {
		fmt.Println("No test results available.")
		return
	}

	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("255")).Bold(true)

	// 准备表头
	headers := []string{
		"Thread",
		"Reqs",
		"Dur(s)",
		"QPS",
		"Toks/s",
		"Avg",
		"P50",
		"P90",
		"P99",
		"1stAvg",
		"1stP50",
		"1stP90",
		"1stP99",
		"ReqToks",
		"ResToks",
	}

	var data [][]string
	// 存储每行对应的并发级别，用于后续样式设置
	concurrencies := make([]int, 0, len(r.concurrentComparison.TestResults))

	for _, result := range r.concurrentComparison.TestResults {
		concurrencies = append(concurrencies, result.Concurrency)

		// 检查是否有失败请求
		reqsCell := fmt.Sprintf("%d", result.Metrics.TotalRequests)
		if result.Metrics.FailedRequests > 0 {
			reqsCell = fmt.Sprintf("%d/%d", result.Metrics.SuccessfulRequests, result.Metrics.TotalRequests)
		}

		row := []string{
			fmt.Sprintf("%d", result.Concurrency),
			reqsCell,
			fmt.Sprintf("%.2f", result.Metrics.TotalDuration.Seconds()),
			fmt.Sprintf("%.2f", result.Metrics.QPS),
			fmt.Sprintf("%.2f", result.Metrics.TokensPerSecond),
			fmt.Sprintf("%d", result.Metrics.AverageLatency.Milliseconds()),
			fmt.Sprintf("%d", result.Metrics.LatencyP50.Milliseconds()),
			fmt.Sprintf("%d", result.Metrics.LatencyP90.Milliseconds()),
			fmt.Sprintf("%d", result.Metrics.LatencyP99.Milliseconds()),
			fmt.Sprintf("%d", result.Metrics.AverageFirstTokenLatency.Milliseconds()),
			fmt.Sprintf("%d", result.Metrics.FirstTokenLatencyP50.Milliseconds()),
			fmt.Sprintf("%d", result.Metrics.FirstTokenLatencyP90.Milliseconds()),
			fmt.Sprintf("%d", result.Metrics.FirstTokenLatencyP99.Milliseconds()),
			fmt.Sprintf("%.1f", result.Metrics.AverageRequestTokens),
			fmt.Sprintf("%.1f", result.Metrics.AverageResponseTokens),
		}
		data = append(data, row)
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(headers...).
		Rows(data...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			if row%2 == 0 {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	fmt.Println(t)
}
