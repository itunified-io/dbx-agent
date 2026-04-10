package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// SessionsCollector collects Oracle session metrics from V$SESSION.
type SessionsCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	enabled    bool
}

func NewSessionsCollector(db OracleQuerier, intervalMs int, systemID string) *SessionsCollector {
	return &SessionsCollector{db: db, intervalMs: intervalMs, systemID: systemID, enabled: true}
}

func (c *SessionsCollector) Name() string           { return "oracle.sessions" }
func (c *SessionsCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *SessionsCollector) Enabled() bool           { return c.enabled }

func (c *SessionsCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID}

	row := c.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END),
			SUM(CASE WHEN blocking_session IS NOT NULL THEN 1 ELSE 0 END),
			(SELECT value FROM v$parameter WHERE name = 'sessions')
		FROM v$session
		WHERE type = 'USER'
	`)

	var total, active, blocked, maxSessions float64
	if err := row.Scan(&total, &active, &blocked, &maxSessions); err != nil {
		return nil, err
	}

	var usagePct float64
	if maxSessions > 0 {
		usagePct = total / maxSessions * 100
	}

	return []collector.Metric{
		{Name: "oracle_session_total", Value: total, Timestamp: now, Labels: labels},
		{Name: "oracle_session_active", Value: active, Timestamp: now, Labels: labels},
		{Name: "oracle_session_blocked", Value: blocked, Timestamp: now, Labels: labels},
		{Name: "oracle_session_max", Value: maxSessions, Timestamp: now, Labels: labels},
		{Name: "oracle_session_usage_pct", Value: usagePct, Timestamp: now, Labels: labels},
	}, nil
}
