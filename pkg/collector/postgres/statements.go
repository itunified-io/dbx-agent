package postgres

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// StatementsCollector collects pg_stat_statements top queries by total time.
type StatementsCollector struct {
	pool       Querier
	intervalMs int
	systemID   string
	topN       int
	enabled    bool
}

func NewStatementsCollector(pool Querier, intervalMs int, systemID string, topN int) *StatementsCollector {
	if topN <= 0 {
		topN = 10
	}
	return &StatementsCollector{pool: pool, intervalMs: intervalMs, systemID: systemID, topN: topN, enabled: true}
}

func (c *StatementsCollector) Name() string           { return "postgres.statements" }
func (c *StatementsCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *StatementsCollector) Enabled() bool           { return c.enabled }

func (c *StatementsCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.pool.Query(ctx,
		`SELECT queryid::text,
		        calls,
		        total_exec_time,
		        mean_exec_time,
		        rows
		 FROM pg_stat_statements
		 ORDER BY total_exec_time DESC
		 LIMIT $1`, c.topN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var queryID string
		var calls int64
		var totalTime, meanTime float64
		var rowCount int64
		if err := rows.Scan(&queryID, &calls, &totalTime, &meanTime, &rowCount); err != nil {
			continue
		}
		labels := map[string]string{"system": c.systemID, "queryid": queryID}
		metrics = append(metrics,
			collector.Metric{Name: "pg_stmt_calls", Value: float64(calls), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_stmt_total_time_ms", Value: totalTime, Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_stmt_mean_time_ms", Value: meanTime, Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_stmt_rows", Value: float64(rowCount), Timestamp: now, Labels: labels},
		)
	}

	return metrics, rows.Err()
}
