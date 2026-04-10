package postgres

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// BuffersCollector collects pg_stat_bgwriter and buffer cache metrics.
type BuffersCollector struct {
	pool       Querier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewBuffersCollector(pool Querier, intervalMs int, systemID string) *BuffersCollector {
	return &BuffersCollector{pool: pool, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *BuffersCollector) Name() string           { return "postgres.buffers" }
func (c *BuffersCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *BuffersCollector) Enabled() bool           { return c.enabled }

func (c *BuffersCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID}
	var metrics []collector.Metric

	// Buffer cache hit ratio
	var blksHit, blksRead float64
	row := c.pool.QueryRow(ctx,
		`SELECT sum(blks_hit), sum(blks_read) FROM pg_stat_database`)
	if err := row.Scan(&blksHit, &blksRead); err != nil {
		return nil, err
	}

	var hitRatio float64
	if blksHit+blksRead > 0 {
		hitRatio = blksHit / (blksHit + blksRead) * 100
	}

	metrics = append(metrics,
		collector.Metric{Name: "pg_buffer_cache_hit_ratio", Value: hitRatio, Timestamp: now, Labels: labels},
		collector.Metric{Name: "pg_buffer_blocks_hit", Value: blksHit, Timestamp: now, Labels: labels},
		collector.Metric{Name: "pg_buffer_blocks_read", Value: blksRead, Timestamp: now, Labels: labels},
	)

	return metrics, nil
}
