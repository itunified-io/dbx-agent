package postgres

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// StatActivityCollector collects pg_stat_activity metrics.
type StatActivityCollector struct {
	pool       Querier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewStatActivityCollector(pool Querier, intervalMs int, systemID string) *StatActivityCollector {
	return &StatActivityCollector{pool: pool, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *StatActivityCollector) Name() string           { return "postgres.stat_activity" }
func (c *StatActivityCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *StatActivityCollector) Enabled() bool           { return c.enabled }

func (c *StatActivityCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.pool.Query(ctx,
		`SELECT COALESCE(state, 'unknown'), count(*)
		 FROM pg_stat_activity
		 WHERE backend_type = 'client backend'
		 GROUP BY state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var state string
		var count int64
		if err := rows.Scan(&state, &count); err != nil {
			continue
		}
		metrics = append(metrics, collector.Metric{
			Name:      "pg_stat_activity_count",
			Value:     float64(count),
			Timestamp: now,
			Labels:    map[string]string{"system": c.systemID, "state": state},
		})
	}

	return metrics, rows.Err()
}
