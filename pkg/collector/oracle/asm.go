package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// ASMCollector collects ASM diskgroup usage metrics.
// License-gated: requires asm option.
type ASMCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	license    LicenseChecker
	enabled    bool
}

func NewASMCollector(db OracleQuerier, intervalMs int, systemID string, lic LicenseChecker) *ASMCollector {
	return &ASMCollector{db: db, intervalMs: intervalMs, systemID: systemID, license: lic, enabled: true}
}

func (c *ASMCollector) Name() string           { return "oracle.asm" }
func (c *ASMCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *ASMCollector) Enabled() bool           { return c.enabled && c.license.IsLicensed("asm") }

func (c *ASMCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.db.QueryContext(ctx, `
		SELECT name, state, type, total_mb, free_mb
		FROM v$asm_diskgroup
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, state, dgType string
		var totalMB, freeMB float64
		if err := rows.Scan(&name, &state, &dgType, &totalMB, &freeMB); err != nil {
			continue
		}
		usedPct := 0.0
		if totalMB > 0 {
			usedPct = (totalMB - freeMB) / totalMB * 100
		}
		labels := map[string]string{"system": c.systemID, "diskgroup": name, "state": state, "type": dgType}
		metrics = append(metrics,
			collector.Metric{Name: "oracle_asm_total_mb", Value: totalMB, Timestamp: now, Labels: labels},
			collector.Metric{Name: "oracle_asm_free_mb", Value: freeMB, Timestamp: now, Labels: labels},
			collector.Metric{Name: "oracle_asm_used_pct", Value: usedPct, Timestamp: now, Labels: labels},
		)
	}

	return metrics, rows.Err()
}
