package cmd

import (
	"encoding/json"

	"github.com/FortuneW/gollmperf/internal/config"
)

// RunFlags holds the command line flags for the run command
type RunFlags struct {
	config.ConfigOverrideFlags
	ConfigPath         string
	IsBatch            bool
	IsPerf             bool
	NoReport           bool
	ShowTableOnConsole bool
	RandomEnable       bool
	RandomEnableSet    bool // true if RandomEnable was explicitly set via command line
	RandomInputLen     int
	RandomOutputLen    int
}

var runFlags = &RunFlags{}

func (r *RunFlags) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}
