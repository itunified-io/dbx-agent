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

// MemoryCollector reads /proc/meminfo for memory usage.
type MemoryCollector struct {
	intervalMs int
	enabled    bool
}

func NewMemoryCollector(intervalMs int) *MemoryCollector {
	return &MemoryCollector{intervalMs: intervalMs, enabled: true}
}

func (c *MemoryCollector) Name() string           { return "host.memory" }
func (c *MemoryCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *MemoryCollector) Enabled() bool           { return c.enabled }

func (c *MemoryCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer f.Close()

	info := make(map[string]uint64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, _ := strconv.ParseUint(parts[1], 10, 64)
		info[key] = val // values in kB
	}

	now := time.Now()
	labels := map[string]string{}
	total := info["MemTotal"]
	available := info["MemAvailable"]
	swapTotal := info["SwapTotal"]
	swapFree := info["SwapFree"]
	hugepagesTotal := info["HugePages_Total"]
	hugepagesFree := info["HugePages_Free"]

	var usedPct float64
	if total > 0 {
		usedPct = float64(total-available) / float64(total) * 100
	}

	swapUsedMB := float64(swapTotal-swapFree) / 1024.0

	return []collector.Metric{
		{Name: "host_memory_total_kb", Value: float64(total), Timestamp: now, Labels: labels},
		{Name: "host_memory_available_kb", Value: float64(available), Timestamp: now, Labels: labels},
		{Name: "host_memory_used_pct", Value: usedPct, Timestamp: now, Labels: labels},
		{Name: "host_memory_swap_used_mb", Value: swapUsedMB, Timestamp: now, Labels: labels},
		{Name: "host_memory_hugepages_total", Value: float64(hugepagesTotal), Timestamp: now, Labels: labels},
		{Name: "host_memory_hugepages_free", Value: float64(hugepagesFree), Timestamp: now, Labels: labels},
	}, nil
}
