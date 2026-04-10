package host

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// ProcessCollector counts running processes by reading /proc.
type ProcessCollector struct {
	intervalMs int
	enabled    bool
}

func NewProcessCollector(intervalMs int) *ProcessCollector {
	return &ProcessCollector{intervalMs: intervalMs, enabled: true}
}

func (c *ProcessCollector) Name() string           { return "host.process" }
func (c *ProcessCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *ProcessCollector) Enabled() bool           { return c.enabled }

func (c *ProcessCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	now := time.Now()
	labels := map[string]string{}
	var total, running, sleeping, zombie int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name[0] < '0' || name[0] > '9' {
			continue
		}
		total++

		statusPath := filepath.Join("/proc", name, "status")
		data, err := os.ReadFile(statusPath)
		if err != nil {
			continue
		}
		for _, line := range splitLines(data) {
			if len(line) > 7 && line[:6] == "State:" {
				state := line[6:]
				for i := range state {
					if state[i] != ' ' && state[i] != '\t' {
						switch state[i] {
						case 'R':
							running++
						case 'S', 'D':
							sleeping++
						case 'Z':
							zombie++
						}
						break
					}
				}
				break
			}
		}
	}

	return []collector.Metric{
		{Name: "host_process_total", Value: float64(total), Timestamp: now, Labels: labels},
		{Name: "host_process_running", Value: float64(running), Timestamp: now, Labels: labels},
		{Name: "host_process_sleeping", Value: float64(sleeping), Timestamp: now, Labels: labels},
		{Name: "host_process_zombie", Value: float64(zombie), Timestamp: now, Labels: labels},
	}, nil
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, string(data[start:i]))
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}
