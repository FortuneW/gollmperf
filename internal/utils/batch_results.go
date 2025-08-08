package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FortuneW/gollmperf/internal/engine"
)

// SaveBatchResultsToJSONL saves batch test results to a JSONL file
// Each line in the output file corresponds to the test case at the same line number in the input file
func SaveBatchResultsToJSONL(results []*engine.Result, filePath string) error {
	_ = os.MkdirAll(filepath.Dir(filePath), 0755)
	// Create or truncate the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create batch results file: %w", err)
	}
	defer file.Close()

	// Write each result as a JSON line
	for i, result := range results {
		var jsonData []byte
		if result.Success && result.RefResponse != nil {
			jsonData = []byte(result.RefResponse.String())
		} else {
			if result.Error != nil {
				jsonData = []byte(result.Error.String())
			}
		}
		// Write to file with newline
		if _, err := file.Write(append(jsonData, '\n')); err != nil {
			return fmt.Errorf("failed to write result %d to file: %w", i, err)
		}
	}

	return nil
}
