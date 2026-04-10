package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// RACCollector collects per-instance metrics from GV$ views.
// License-gated: requires rac option.
type RACCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	license    LicenseChecker
	enabled    bool
}

func NewRACCollector(db OracleQuerier, intervalMs int, systemID string, lic LicenseChecker) *RACCollector {
	return &RACCollector{db: db, intervalMs: intervalMs, systemID: systemID, license: lic, enabled: true}
}

func (c *RACCollector) Name() string           { return "oracle.rac" }
func (c *RACCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *RACCollector) Enabled() bool           { return c.enabled && c.license.IsLicensed("rac") }

func (c *RACCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.db.QueryContext(ctx, `
		SELECT inst_id, instance_name, status,
		       (SELECT COUNT(*) FROM gv$session s WHERE s.inst_id = i.inst_id AND s.type = 'USER') AS user_sessions
		FROM gv$instance i
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var instID int
		var name, status string
		var sessions int64
		if err := rows.Scan(&instID, &name, &status, &sessions); err != nil {
			continue
		}
		labels := map[string]string{"system": c.systemID, "instance": name, "status": status}
		metrics = append(metrics, collector.Metric{
			Name: "oracle_rac_instance_sessions", Value: float64(sessions), Timestamp: now, Labels: labels,
		})
	}

	return metrics, rows.Err()
}
