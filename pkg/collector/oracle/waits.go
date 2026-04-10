package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// WaitsCollector collects top wait events from V$SYSTEM_EVENT.
type WaitsCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewWaitsCollector(db OracleQuerier, intervalMs int, systemID string) *WaitsCollector {
	return &WaitsCollector{db: db, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *WaitsCollector) Name() string           { return "oracle.waits" }
func (c *WaitsCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *WaitsCollector) Enabled() bool           { return c.enabled }

func (c *WaitsCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.db.QueryContext(ctx, `
		SELECT event, wait_class, time_waited_micro / 1000 AS time_waited_ms, total_waits
		FROM v$system_event
		WHERE wait_class NOT IN ('Idle')
		ORDER BY time_waited_micro DESC
		FETCH FIRST 20 ROWS ONLY
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event, waitClass string
		var timeWaitedMs float64
		var totalWaits int64
		if err := rows.Scan(&event, &waitClass, &timeWaitedMs, &totalWaits); err != nil {
			continue
		}
		labels := map[string]string{"system": c.systemID, "event": event, "wait_class": waitClass}
		metrics = append(metrics,
			collector.Metric{Name: "oracle_wait_time_ms", Value: timeWaitedMs, Timestamp: now, Labels: labels},
			collector.Metric{Name: "oracle_wait_total", Value: float64(totalWaits), Timestamp: now, Labels: labels},
		)
	}

	return metrics, rows.Err()
}
