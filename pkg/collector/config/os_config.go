package config

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// OSConfigCollector captures OS-level configuration snapshots.
type OSConfigCollector struct {
	intervalMs int
	enabled    bool
}

func NewOSConfigCollector(intervalMs int) *OSConfigCollector {
	return &OSConfigCollector{intervalMs: intervalMs, enabled: true}
}

func (c *OSConfigCollector) Name() string           { return "config.os" }
func (c *OSConfigCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *OSConfigCollector) Enabled() bool           { return c.enabled }

func (c *OSConfigCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	now := time.Now()
	labels := map[string]string{"snapshot_type": "os_config"}
	var metrics []collector.Metric

	hostname, _ := os.Hostname()
	labels["hostname"] = hostname
	labels["os"] = runtime.GOOS
	labels["arch"] = runtime.GOARCH

	metrics = append(metrics, collector.Metric{
		Name: "config_os_num_cpu", Value: float64(runtime.NumCPU()), Timestamp: now, Labels: labels,
	})

	// Kernel version on Linux
	if data, err := os.ReadFile("/proc/version"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) >= 3 {
			labels["kernel"] = parts[2]
		}
	}

	return metrics, nil
}
