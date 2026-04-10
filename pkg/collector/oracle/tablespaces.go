package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// TablespacesCollector collects per-tablespace usage from DBA_TABLESPACE_USAGE_METRICS.
type TablespacesCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewTablespacesCollector(db OracleQuerier, intervalMs int, systemID string) *TablespacesCollector {
	return &TablespacesCollector{db: db, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *TablespacesCollector) Name() string           { return "oracle.tablespaces" }
func (c *TablespacesCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *TablespacesCollector) Enabled() bool           { return c.enabled }

func (c *TablespacesCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.db.QueryContext(ctx, `
		SELECT tablespace_name, ROUND(used_percent, 2)
		FROM dba_tablespace_usage_metrics
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var usedPct float64
		if err := rows.Scan(&name, &usedPct); err != nil {
			continue
		}
		metrics = append(metrics, collector.Metric{
			Name:      "oracle_tablespace_used_pct",
			Value:     usedPct,
			Timestamp: now,
			Labels:    map[string]string{"system": c.systemID, "tablespace": name},
		})
	}

	return metrics, rows.Err()
}
