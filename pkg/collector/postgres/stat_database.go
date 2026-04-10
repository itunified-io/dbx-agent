package postgres

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// StatDatabaseCollector collects pg_stat_database per-database metrics.
type StatDatabaseCollector struct {
	pool       Querier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewStatDatabaseCollector(pool Querier, intervalMs int, systemID string) *StatDatabaseCollector {
	return &StatDatabaseCollector{pool: pool, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *StatDatabaseCollector) Name() string           { return "postgres.stat_database" }
func (c *StatDatabaseCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *StatDatabaseCollector) Enabled() bool           { return c.enabled }

func (c *StatDatabaseCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	var metrics []collector.Metric

	rows, err := c.pool.Query(ctx,
		`SELECT datname,
		        numbackends,
		        xact_commit,
		        xact_rollback,
		        tup_returned,
		        tup_fetched,
		        tup_inserted,
		        tup_updated,
		        tup_deleted,
		        deadlocks
		 FROM pg_stat_database
		 WHERE datname IS NOT NULL AND datname NOT IN ('template0', 'template1')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dbName string
		var backends int64
		var commits, rollbacks, returned, fetched, inserted, updated, deleted, deadlocks int64
		if err := rows.Scan(&dbName, &backends, &commits, &rollbacks, &returned, &fetched, &inserted, &updated, &deleted, &deadlocks); err != nil {
			continue
		}
		labels := map[string]string{"system": c.systemID, "database": dbName}
		metrics = append(metrics,
			collector.Metric{Name: "pg_db_backends", Value: float64(backends), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_xact_commit", Value: float64(commits), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_xact_rollback", Value: float64(rollbacks), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_tup_returned", Value: float64(returned), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_tup_fetched", Value: float64(fetched), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_tup_inserted", Value: float64(inserted), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_tup_updated", Value: float64(updated), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_tup_deleted", Value: float64(deleted), Timestamp: now, Labels: labels},
			collector.Metric{Name: "pg_db_deadlocks", Value: float64(deadlocks), Timestamp: now, Labels: labels},
		)
	}

	return metrics, rows.Err()
}
