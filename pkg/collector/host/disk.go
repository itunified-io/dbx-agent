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

// DiskCollector reads /proc/diskstats for disk I/O metrics.
type DiskCollector struct {
	intervalMs int
	enabled    bool
	prevStats  map[string]diskStat
}

type diskStat struct {
	readsCompleted  uint64
	writesCompleted uint64
	readBytes       uint64
	writeBytes      uint64
	ioTime          uint64
}

func NewDiskCollector(intervalMs int) *DiskCollector {
	return &DiskCollector{intervalMs: intervalMs, enabled: true, prevStats: make(map[string]diskStat)}
}

func (c *DiskCollector) Name() string           { return "host.disk" }
func (c *DiskCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *DiskCollector) Enabled() bool           { return c.enabled }

func (c *DiskCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	f, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, fmt.Errorf("open /proc/diskstats: %w", err)
	}
	defer f.Close()

	now := time.Now()
	var metrics []collector.Metric
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}
		dev := fields[2]
		// Skip partitions (only collect whole disks like sda, vda, nvme0n1)
		if strings.HasPrefix(dev, "loop") || strings.HasPrefix(dev, "ram") || strings.HasPrefix(dev, "dm-") {
			continue
		}

		reads, _ := strconv.ParseUint(fields[3], 10, 64)
		readSectors, _ := strconv.ParseUint(fields[5], 10, 64)
		writes, _ := strconv.ParseUint(fields[7], 10, 64)
		writeSectors, _ := strconv.ParseUint(fields[9], 10, 64)
		ioTime, _ := strconv.ParseUint(fields[12], 10, 64)

		labels := map[string]string{"device": dev}

		current := diskStat{
			readsCompleted:  reads,
			writesCompleted: writes,
			readBytes:       readSectors * 512,
			writeBytes:      writeSectors * 512,
			ioTime:          ioTime,
		}

		if prev, ok := c.prevStats[dev]; ok {
			metrics = append(metrics,
				collector.Metric{Name: "host_disk_reads_per_sec", Value: float64(current.readsCompleted - prev.readsCompleted), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_disk_writes_per_sec", Value: float64(current.writesCompleted - prev.writesCompleted), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_disk_read_bytes", Value: float64(current.readBytes - prev.readBytes), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_disk_write_bytes", Value: float64(current.writeBytes - prev.writeBytes), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_disk_io_time_ms", Value: float64(current.ioTime - prev.ioTime), Timestamp: now, Labels: labels},
			)
		}

		c.prevStats[dev] = current
	}

	return metrics, nil
}
