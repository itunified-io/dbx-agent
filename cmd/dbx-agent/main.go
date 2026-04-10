package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/itunified-io/dbx-agent/pkg/agent"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:   "dbx-agent",
		Short: "dbx monitoring agent — collects database and host metrics",
		Long:  `Lightweight Go agent that collects metrics from Oracle, PostgreSQL, and OS targets, sending them to VictoriaMetrics and dbx-central.`,
		RunE:  runAgent,
	}

	root.PersistentFlags().String("config", "", "config file path (default: /opt/dbx-agent/config/agent.yaml or $DBX_AGENT_CONFIG)")
	root.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")

	root.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print agent version",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("dbx-agent %s\n", version)
			},
		},
		&cobra.Command{
			Use:   "check-config",
			Short: "Validate config file without starting the agent",
			RunE: func(cmd *cobra.Command, args []string) error {
				path := configPath(cmd)
				_, err := agent.LoadConfig(path)
				if err != nil {
					return fmt.Errorf("config validation failed: %w", err)
				}
				fmt.Printf("Config OK: %s\n", path)
				return nil
			},
		},
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runAgent(cmd *cobra.Command, args []string) error {
	level := slog.LevelInfo
	if lvl, _ := cmd.Flags().GetString("log-level"); lvl != "" {
		switch lvl {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	path := configPath(cmd)
	cfg, err := agent.LoadConfig(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	a := agent.NewAgent(cfg, version)
	return a.Run(context.Background())
}

func configPath(cmd *cobra.Command) string {
	if path, _ := cmd.Flags().GetString("config"); path != "" {
		return path
	}
	if v := os.Getenv("DBX_AGENT_CONFIG"); v != "" {
		return v
	}
	return "/opt/dbx-agent/config/agent.yaml"
}
