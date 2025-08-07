package config

import (
	"encoding/json"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Load the example config file
	config, err := LoadConfig("../../configs/example.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	b, _ := json.MarshalIndent(config, "", "  ")
	t.Logf("Full config: %s", string(b))

}
