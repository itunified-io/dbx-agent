package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
	"github.com/itunified-io/dbx-agent/pkg/sink"
)

type Scheduler struct {
	registry *collector.Registry
	sinks    []sink.Sink
}

func NewScheduler(reg *collector.Registry, sinks []sink.Sink) *Scheduler {
	return &Scheduler{registry: reg, sinks: sinks}
}

func (s *Scheduler) Run(ctx context.Context) {
	for _, c := range s.registry.Enabled() {
		go s.runCollector(ctx, c)
	}
	<-ctx.Done()
}

func (s *Scheduler) runCollector(ctx context.Context, c collector.Collector) {
	ticker := time.NewTicker(c.Interval())
	defer ticker.Stop()
	s.collect(ctx, c)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collect(ctx, c)
		}
	}
}

func (s *Scheduler) collect(ctx context.Context, c collector.Collector) {
	start := time.Now()
	metrics, err := c.Collect(ctx)
	duration := time.Since(start)
	if err != nil {
		slog.Error("collector failed", "collector", c.Name(), "error", err, "duration", duration)
		return
	}
	slog.Debug("collector completed", "collector", c.Name(), "metrics", len(metrics), "duration", duration)
	for _, sk := range s.sinks {
		if err := sk.Send(ctx, metrics); err != nil {
			slog.Error("sink send failed", "sink", sk.Name(), "collector", c.Name(), "error", err)
		}
	}
}
