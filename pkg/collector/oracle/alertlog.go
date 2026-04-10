package oracle

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// AlertLogCollector tails the Oracle alert log for ORA- errors.
type AlertLogCollector struct {
	alertLogPath string
	intervalMs   int
	systemID     string
	enabled      bool
	lastOffset   int64
}

func NewAlertLogCollector(alertLogPath string, intervalMs int, systemID string) *AlertLogCollector {
	return &AlertLogCollector{
		alertLogPath: alertLogPath,
		intervalMs:   intervalMs,
		systemID:     systemID,
		enabled:      true,
	}
}

func (c *AlertLogCollector) Name() string           { return "oracle.alertlog" }
func (c *AlertLogCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *AlertLogCollector) Enabled() bool           { return c.enabled }

func (c *AlertLogCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"system": c.systemID}

	f, err := os.Open(c.alertLogPath)
	if err != nil {
		return nil, fmt.Errorf("open alert log %s: %w", c.alertLogPath, err)
	}
	defer f.Close()

	// Seek to last known position
	if c.lastOffset > 0 {
		f.Seek(c.lastOffset, 0)
	}

	scanner := bufio.NewScanner(f)
	var oraErrors int
	severityCounts := map[string]int{
		"critical": 0,
		"warning":  0,
		"info":     0,
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ORA-") {
			oraErrors++
			severity := classifyORAError(line)
			severityCounts[severity]++
		}
	}

	// Update offset
	offset, _ := f.Seek(0, 1)
	c.lastOffset = offset

	metrics := []collector.Metric{
		{Name: "oracle_alertlog_ora_errors", Value: float64(oraErrors), Timestamp: now, Labels: labels},
	}
	for sev, count := range severityCounts {
		metrics = append(metrics, collector.Metric{
			Name:      "oracle_alertlog_severity_count",
			Value:     float64(count),
			Timestamp: now,
			Labels:    map[string]string{"system": c.systemID, "severity": sev},
		})
	}

	return metrics, nil
}

func classifyORAError(line string) string {
	// Critical: ORA-00600 (internal error), ORA-07445 (exception), ORA-04031 (shared pool)
	critical := []string{"ORA-00600", "ORA-07445", "ORA-04031", "ORA-01578", "ORA-01110"}
	for _, code := range critical {
		if strings.Contains(line, code) {
			return "critical"
		}
	}
	// Warning: ORA-01555 (snapshot too old), ORA-12154 (TNS), ORA-28001 (password expired)
	warning := []string{"ORA-01555", "ORA-12154", "ORA-28001", "ORA-01652", "ORA-30036"}
	for _, code := range warning {
		if strings.Contains(line, code) {
			return "warning"
		}
	}
	return "info"
}
