// Prepare data for the chart
const testData = [
{{- range .ReporterData.TestResults }}
  {
    concurrency: {{.Concurrency}},
    qps: {{printf "%.2f" .Metrics.QPS}},
    tokensPerSec: {{printf "%.1f" .Metrics.TokensPerSecond}}
  },
{{- end }}
];

const concurrencyLevels = testData.map(item => item.concurrency);
const qpsValues = testData.map(item => item.qps);
const tokensPerSecValues = testData.map(item => item.tokensPerSec);

// Create the chart
const ctx = document.getElementById('performanceChart').getContext('2d');
new Chart(ctx, {
    type: 'line',
    data: {
        labels: concurrencyLevels,
        datasets: [
            {
                label: 'QPS (Queries Per Second)',
                data: qpsValues,
                borderColor: '#2196f3',
                backgroundColor: 'rgba(33, 150, 243, 0.1)',
                borderWidth: 2,
                fill: false,
                yAxisID: 'y'
            },
            {
                label: 'Tokens/sec',
                data: tokensPerSecValues,
                borderColor: '#4caf50',
                backgroundColor: 'rgba(76, 175, 80, 0.1)',
                borderWidth: 2,
                fill: false,
                yAxisID: 'y1'
            }
        ]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
            mode: 'index',
            intersect: false
        },
        scales: {
            x: {
                title: {
                    display: true,
                    text: 'Concurrency Level'
                }
            },
            y: {
                type: 'linear',
                display: true,
                position: 'left',
                title: {
                    display: true,
                    text: 'QPS'
                }
            },
            y1: {
                type: 'linear',
                display: true,
                position: 'right',
                title: {
                    display: true,
                    text: 'Tokens/sec'
                },
                grid: {
                    drawOnChartArea: false
                }
            }
        },
        plugins: {
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: function (context) {
                        let label = context.dataset.label || '';
                        if (label) {
                            label += ': ';
                        }
                        if (context.parsed.y !== null) {
                            label += context.parsed.y.toFixed(2);
                        }
                        return label;
                    }
                }
            }
        }
    }
});