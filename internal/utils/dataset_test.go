package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadDataset
func TestLoadDataset(t *testing.T) {
	content := `
	{"messages":[{"role":"user","content":"Test message 1"}]}
{"messages":[{"role":"user","content":"Test message 2"}]}
`
	file, err := os.CreateTemp("", "test_dataset_*.jsonl")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString(content)
	assert.NoError(t, err)
	file.Close()

	dataset, err := LoadDataset(file.Name(), "jsonl")
	assert.NoError(t, err)
	assert.Len(t, dataset, 2)

	t.Log(dataset)
}

func TestLoadExamplesCaseFile(t *testing.T) {
	dataset, err := LoadDataset("../../examples/test_cases.jsonl", "jsonl")
	assert.NoError(t, err)
	t.Log(dataset)
}
