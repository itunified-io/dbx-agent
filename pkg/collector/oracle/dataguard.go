package oracle

import (
	"context"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// DataGuardCollector collects Data Guard transport/apply lag metrics.
// License-gated: requires diagnostics_pack.
type DataGuardCollector struct {
	db         OracleQuerier
	intervalMs int
	systemID   string
	license    LicenseChecker
	enabled    bool
}

func NewDataGuardCollector(db OracleQuerier, intervalMs int, systemID string, lic LicenseChecker) *DataGuardCollector {
	return &DataGuardCollector{db: db, intervalMs: intervalMs, systemID: systemID, license: lic, enabled: true}
}

func (c *DataGuardCollector) Name() string           { return "oracle.dataguard" }
func (c *DataGuardCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *DataGuardCollector) Enabled() bool           { return c.enabled && c.license.IsLicensed("diagnostics_pack") }

func (c *DataGuardCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID}
	var metrics []collector.Metric

	rows, err := c.db.QueryContext(ctx, `
		SELECT name, value FROM v$dataguard_stats
		WHERE name IN ('transport lag', 'apply lag', 'apply finish time')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			continue
		}
		metrics = append(metrics, collector.Metric{
			Name:      "oracle_dg_" + sanitizeName(name),
			Value:     parseIntervalToSeconds(value),
			Timestamp: now,
			Labels:    labels,
		})
	}

	return metrics, rows.Err()
}

func sanitizeName(s string) string {
	out := make([]byte, 0, len(s))
	for _, b := range []byte(s) {
		if b == ' ' {
			out = append(out, '_')
		} else if (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '_' {
			out = append(out, b)
		} else if b >= 'A' && b <= 'Z' {
			out = append(out, b+32) // lowercase
		}
	}
	return string(out)
}

// parseIntervalToSeconds converts Oracle interval strings like "+00 00:00:05" to seconds.
func parseIntervalToSeconds(s string) float64 {
	// Format: +DD HH:MI:SS or similar
	if len(s) == 0 {
		return 0
	}
	var days, hours, minutes, seconds float64
	// Try parsing "+DD HH:MM:SS" format
	n, _ := parseIntervalComponents(s)
	_ = n
	days = 0
	hours = 0
	minutes = 0
	seconds = 0

	// Simple parser for "+DD HH:MI:SS"
	parts := splitInterval(s)
	if len(parts) >= 4 {
		days = parseFloat(parts[0])
		hours = parseFloat(parts[1])
		minutes = parseFloat(parts[2])
		seconds = parseFloat(parts[3])
	}
	return days*86400 + hours*3600 + minutes*60 + seconds
}

func splitInterval(s string) []string {
	// Remove leading +/-
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		s = s[1:]
	}
	// Split on space and colon
	var parts []string
	current := ""
	for _, c := range s {
		if c == ' ' || c == ':' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func parseFloat(s string) float64 {
	var v float64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + float64(c-'0')
		} else if c == '.' {
			// Simple integer parse is sufficient
			break
		}
	}
	return v
}

func parseIntervalComponents(s string) (int, error) {
	return 0, nil
}
