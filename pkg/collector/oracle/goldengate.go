package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// GoldenGateCollector collects GoldenGate process status and lag.
// License-gated: requires goldengate option.
type GoldenGateCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	license    LicenseChecker
	enabled    bool
}

func NewGoldenGateCollector(db OracleQuerier, intervalMs int, systemID string, lic LicenseChecker) *GoldenGateCollector {
	return &GoldenGateCollector{db: db, intervalMs: intervalMs, systemID: systemID, license: lic, enabled: true}
}

func (c *GoldenGateCollector) Name() string           { return "oracle.goldengate" }
func (c *GoldenGateCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *GoldenGateCollector) Enabled() bool           { return c.enabled && c.license.IsLicensed("goldengate") }

func (c *GoldenGateCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.db.QueryContext(ctx, `
		SELECT program_name, program_status,
		       COALESCE(EXTRACT(SECOND FROM lag), 0) +
		       COALESCE(EXTRACT(MINUTE FROM lag), 0) * 60 +
		       COALESCE(EXTRACT(HOUR FROM lag), 0) * 3600 AS lag_seconds
		FROM dba_goldengate_inbound
		UNION ALL
		SELECT capture_name, status,
		       COALESCE(EXTRACT(SECOND FROM lag), 0) +
		       COALESCE(EXTRACT(MINUTE FROM lag), 0) * 60 +
		       COALESCE(EXTRACT(HOUR FROM lag), 0) * 3600
		FROM dba_goldengate_outbound
	`)
	if err != nil {
		// GoldenGate views may not exist — not an error, just no metrics
		return metrics, nil
	}
	defer rows.Close()

	for rows.Next() {
		var name, status string
		var lagSeconds float64
		if err := rows.Scan(&name, &status, &lagSeconds); err != nil {
			continue
		}
		labels := map[string]string{"system": c.systemID, "process": name, "status": status}
		metrics = append(metrics,
			collector.Metric{Name: "oracle_gg_lag_seconds", Value: lagSeconds, Timestamp: now, Labels: labels},
		)
	}

	return metrics, rows.Err()
}
