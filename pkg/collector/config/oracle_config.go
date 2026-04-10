package config

import (
	"context"
	"database/sql"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// OracleConfigCollector captures Oracle configuration snapshots.
type OracleConfigCollector struct {
	db         *sql.DB
	intervalMs int
	systemID   string
	enabled    bool
}

func NewOracleConfigCollector(db *sql.DB, intervalMs int, systemID string) *OracleConfigCollector {
	return &OracleConfigCollector{db: db, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *OracleConfigCollector) Name() string           { return "config.oracle" }
func (c *OracleConfigCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *OracleConfigCollector) Enabled() bool           { return c.enabled }

func (c *OracleConfigCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID, "snapshot_type": "oracle_config"}
	var metrics []collector.Metric

	// Count non-default parameters
	if c.db != nil {
		row := c.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM v$parameter WHERE isdefault = 'FALSE'`)
		var count float64
		if err := row.Scan(&count); err == nil {
			metrics = append(metrics, collector.Metric{
				Name: "config_oracle_modified_params", Value: count, Timestamp: now, Labels: labels,
			})
		}
	}

	return metrics, nil
}
