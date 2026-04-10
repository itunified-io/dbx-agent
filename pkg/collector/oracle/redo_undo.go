package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// RedoUndoCollector collects redo log and undo metrics.
type RedoUndoCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewRedoUndoCollector(db OracleQuerier, intervalMs int, systemID string) *RedoUndoCollector {
	return &RedoUndoCollector{db: db, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *RedoUndoCollector) Name() string           { return "oracle.redo_undo" }
func (c *RedoUndoCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *RedoUndoCollector) Enabled() bool           { return c.enabled }

func (c *RedoUndoCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID}
	var metrics []collector.Metric

	// Redo log switches per hour
	var switchesPerHour float64
	row := c.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM v$log_history
		WHERE first_time > SYSDATE - 1/24
	`)
	if err := row.Scan(&switchesPerHour); err != nil {
		return nil, err
	}
	metrics = append(metrics, collector.Metric{
		Name: "oracle_redo_switches_per_hour", Value: switchesPerHour, Timestamp: now, Labels: labels,
	})

	// Undo retention and usage
	var undoRetention float64
	row = c.db.QueryRowContext(ctx,
		`SELECT value FROM v$parameter WHERE name = 'undo_retention'`)
	if err := row.Scan(&undoRetention); err != nil {
		return nil, err
	}
	metrics = append(metrics, collector.Metric{
		Name: "oracle_undo_retention_seconds", Value: undoRetention, Timestamp: now, Labels: labels,
	})

	return metrics, nil
}
