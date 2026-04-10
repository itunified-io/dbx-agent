package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// SGAPGACollector collects SGA and PGA memory metrics.
type SGAPGACollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewSGAPGACollector(db OracleQuerier, intervalMs int, systemID string) *SGAPGACollector {
	return &SGAPGACollector{db: db, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *SGAPGACollector) Name() string           { return "oracle.sga_pga" }
func (c *SGAPGACollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *SGAPGACollector) Enabled() bool           { return c.enabled }

func (c *SGAPGACollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID}
	var metrics []collector.Metric

	// SGA size
	var sgaSize float64
	row := c.db.QueryRowContext(ctx, `SELECT SUM(value) FROM v$sga`)
	if err := row.Scan(&sgaSize); err != nil {
		return nil, err
	}
	metrics = append(metrics, collector.Metric{
		Name: "oracle_sga_total_bytes", Value: sgaSize, Timestamp: now, Labels: labels,
	})

	// PGA allocated
	var pgaAlloc float64
	row = c.db.QueryRowContext(ctx,
		`SELECT value FROM v$pgastat WHERE name = 'total PGA allocated'`)
	if err := row.Scan(&pgaAlloc); err != nil {
		return nil, err
	}
	metrics = append(metrics, collector.Metric{
		Name: "oracle_pga_allocated_bytes", Value: pgaAlloc, Timestamp: now, Labels: labels,
	})

	// Buffer cache hit ratio
	var hitRatio float64
	row = c.db.QueryRowContext(ctx,
		`SELECT ROUND((1 - (phy.value / (cur.value + con.value))) * 100, 2)
		 FROM v$sysstat phy, v$sysstat cur, v$sysstat con
		 WHERE phy.name = 'physical reads'
		   AND cur.name = 'db block gets'
		   AND con.name = 'consistent gets'`)
	if err := row.Scan(&hitRatio); err != nil {
		return nil, err
	}
	metrics = append(metrics, collector.Metric{
		Name: "oracle_buffer_cache_hit_ratio", Value: hitRatio, Timestamp: now, Labels: labels,
	})

	return metrics, nil
}
