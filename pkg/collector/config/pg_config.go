package config

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
	pgcol "github.com/itunified-io/dbx-agent/pkg/collector/postgres"
)

// PGConfigCollector captures PostgreSQL configuration snapshots.
type PGConfigCollector struct {
	pool       pgcol.Querier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewPGConfigCollector(pool pgcol.Querier, intervalMs int, systemID string) *PGConfigCollector {
	return &PGConfigCollector{pool: pool, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *PGConfigCollector) Name() string           { return "config.postgres" }
func (c *PGConfigCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *PGConfigCollector) Enabled() bool           { return c.enabled }

func (c *PGConfigCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID, "snapshot_type": "pg_config"}
	var metrics []collector.Metric

	if c.pool != nil {
		// Count non-default settings
		row := c.pool.QueryRow(ctx,
			`SELECT count(*) FROM pg_settings WHERE source != 'default'`)
		var count int64
		if err := row.Scan(&count); err == nil {
			metrics = append(metrics, collector.Metric{
				Name: "config_pg_modified_params", Value: float64(count), Timestamp: now, Labels: labels,
			})
		}
	}

	return metrics, nil
}
