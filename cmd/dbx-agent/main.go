package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/itunified-io/dbx-agent/pkg/agent"
)

var version = "dev"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("dbx-agent %s\n", version)
		os.Exit(0)
	}

	configPath := "/opt/dbx-agent/config/agent.yaml"
	if v := os.Getenv("DBX_AGENT_CONFIG"); v != "" {
		configPath = v
	}

	cfg, err := agent.LoadConfig(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	a := agent.NewAgent(cfg, version)
	if err := a.Run(context.Background()); err != nil {
		slog.Error("agent failed", "error", err)
		os.Exit(1)
	}
}
