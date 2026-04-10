package host

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// CPUCollector reads /proc/stat to calculate CPU usage percentages.
type CPUCollector struct {
	intervalMs int
	enabled    bool
	prevIdle   uint64
	prevTotal  uint64
}

// NewCPUCollector creates a CPU collector with the given interval in milliseconds.
func NewCPUCollector(intervalMs int) *CPUCollector {
	return &CPUCollector{intervalMs: intervalMs, enabled: true}
}

func (c *CPUCollector) Name() string           { return "host.cpu" }
func (c *CPUCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *CPUCollector) Enabled() bool           { return c.enabled }

func (c *CPUCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return nil, fmt.Errorf("open /proc/stat: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	now := time.Now()
	var metrics []collector.Metric

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 8 {
			break
		}

		var values [7]uint64
		for i := 1; i <= 7 && i < len(fields); i++ {
			values[i-1], _ = strconv.ParseUint(fields[i], 10, 64)
		}

		// user, nice, system, idle, iowait, irq, softirq
		idle := values[3] + values[4]
		total := values[0] + values[1] + values[2] + values[3] + values[4] + values[5] + values[6]

		if c.prevTotal > 0 {
			deltaTotal := total - c.prevTotal
			deltaIdle := idle - c.prevIdle
			if deltaTotal > 0 {
				usagePct := float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
				iowaitPct := float64(values[4]) / float64(deltaTotal) * 100

				metrics = append(metrics,
					collector.Metric{Name: "host_cpu_usage_pct", Value: usagePct, Timestamp: now, Labels: map[string]string{}},
					collector.Metric{Name: "host_cpu_iowait_pct", Value: iowaitPct, Timestamp: now, Labels: map[string]string{}},
				)
			}
		}

		c.prevIdle = idle
		c.prevTotal = total
		break
	}

	if len(metrics) == 0 {
		// First read: return zero values (delta requires two reads)
		metrics = append(metrics,
			collector.Metric{Name: "host_cpu_usage_pct", Value: 0, Timestamp: now, Labels: map[string]string{}},
			collector.Metric{Name: "host_cpu_iowait_pct", Value: 0, Timestamp: now, Labels: map[string]string{}},
		)
	}

	return metrics, nil
}
