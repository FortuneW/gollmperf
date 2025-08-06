package main

import (
	"fmt"
	"os"

	"github.com/FortuneW/qlog"
	"github.com/user/llmperf/cmd"
)

func init() {
	qlog.InitWithConfig(qlog.Config{
		Level:        "DEB",
		Mode:         "console",
		ToConsole:    true,
		ColorConsole: true,
	})
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
