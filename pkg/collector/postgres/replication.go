package postgres

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// ReplicationCollector collects pg_stat_replication metrics.
type ReplicationCollector struct {
	pool       Querier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewReplicationCollector(pool Querier, intervalMs int, systemID string) *ReplicationCollector {
	return &ReplicationCollector{pool: pool, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *ReplicationCollector) Name() string           { return "postgres.replication" }
func (c *ReplicationCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *ReplicationCollector) Enabled() bool           { return c.enabled }

func (c *ReplicationCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.pool.Query(ctx,
		`SELECT client_addr::text,
		        COALESCE(state, 'unknown'),
		        pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replay_lag_bytes,
		        COALESCE(EXTRACT(EPOCH FROM replay_lag), 0) AS replay_lag_seconds
		 FROM pg_stat_replication`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var addr, state string
		var lagBytes, lagSeconds float64
		if err := rows.Scan(&addr, &state, &lagBytes, &lagSeconds); err != nil {
			continue
		}
		labels := map[string]string{"system": c.systemID, "client": addr, "state": state}
		metrics = append(metrics,
			collector.Metric{Name: "pg_replication_lag_bytes", Value: lagBytes, Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_replication_lag_seconds", Value: lagSeconds, Timestamp: now, Labels: labels},
		)
	}

	return metrics, rows.Err()
}
