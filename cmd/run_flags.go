package cmd

import (
	"encoding/json"

	"github.com/user/llmperf/internal/config"
)

// RunFlags holds the command line flags for the run command
type RunFlags struct {
	config.ConfigOverrideFlags
	ConfigPath string
	IsStress   bool
}

var runFlags = &RunFlags{}

func (r *RunFlags) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}
