package postgres

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// LocksCollector collects pg_locks metrics including blocking chains.
type LocksCollector struct {
	pool       Querier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewLocksCollector(pool Querier, intervalMs int, systemID string) *LocksCollector {
	return &LocksCollector{pool: pool, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *LocksCollector) Name() string           { return "postgres.locks" }
func (c *LocksCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *LocksCollector) Enabled() bool           { return c.enabled }

func (c *LocksCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID}
	var metrics []collector.Metric

	// Lock count by mode
	rows, err := c.pool.Query(ctx,
		`SELECT mode, count(*)
		 FROM pg_locks
		 WHERE granted = true
		 GROUP BY mode`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mode string
		var count int64
		if err := rows.Scan(&mode, &count); err != nil {
			continue
		}
		metrics = append(metrics, collector.Metric{
			Name:      "pg_locks_count",
			Value:     float64(count),
			Timestamp: now,
			Labels:    map[string]string{"system": c.systemID, "mode": mode},
		})
	}
	if err := rows.Err(); err != nil {
		return metrics, err
	}

	// Blocked queries count
	var blocked int64
	row := c.pool.QueryRow(ctx,
		`SELECT count(*) FROM pg_locks WHERE NOT granted`)
	if err := row.Scan(&blocked); err != nil {
		return metrics, err
	}
	metrics = append(metrics, collector.Metric{
		Name: "pg_locks_blocked", Value: float64(blocked), Timestamp: now, Labels: labels,
	})

	return metrics, nil
}
