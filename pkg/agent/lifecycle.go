package agent

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Agent struct {
	config    *AgentConfig
	scheduler *Scheduler
	health    *HealthServer
	httpSrv   *http.Server
	version   string
}

func NewAgent(cfg *AgentConfig, version string) *Agent {
	return &Agent{
		config:  cfg,
		health:  NewHealthServer(),
		version: version,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	a.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Agent.HostPort),
		Handler: a.health.Handler(a.config.Agent.ID, a.version),
	}
	go func() {
		slog.Info("health server starting", "port", a.config.Agent.HostPort)
		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("health server failed", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	slog.Info("dbx-agent started", "id", a.config.Agent.ID, "version", a.version)

	for {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGHUP:
				slog.Info("SIGHUP received, reloading config")
			case syscall.SIGTERM, syscall.SIGINT:
				slog.Info("shutdown signal received", "signal", sig)
				return a.shutdown()
			}
		case <-ctx.Done():
			return a.shutdown()
		}
	}
}

func (a *Agent) shutdown() error {
	slog.Info("shutting down dbx-agent")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if a.httpSrv != nil {
		a.httpSrv.Shutdown(shutdownCtx)
	}
	slog.Info("dbx-agent stopped")
	return nil
}
