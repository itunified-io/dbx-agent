package host

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// FilesystemCollector reads mounted filesystems and reports usage via statfs.
type FilesystemCollector struct {
	intervalMs int
	enabled    bool
}

func NewFilesystemCollector(intervalMs int) *FilesystemCollector {
	return &FilesystemCollector{intervalMs: intervalMs, enabled: true}
}

func (c *FilesystemCollector) Name() string           { return "host.filesystem" }
func (c *FilesystemCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *FilesystemCollector) Enabled() bool           { return c.enabled }

func (c *FilesystemCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	mounts, err := parseMounts()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var metrics []collector.Metric

	for _, m := range mounts {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(m.mountpoint, &stat); err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		avail := stat.Bavail * uint64(stat.Bsize)
		free := stat.Bfree * uint64(stat.Bsize)
		used := total - free
		if total == 0 {
			continue
		}

		usedPct := float64(used) / float64(total) * 100
		labels := map[string]string{"mount": m.mountpoint, "device": m.device, "fstype": m.fstype}

		metrics = append(metrics,
			collector.Metric{Name: "host_fs_total_bytes", Value: float64(total), Timestamp: now, Labels: labels},
			collector.Metric{Name: "host_fs_used_bytes", Value: float64(used), Timestamp: now, Labels: labels},
			collector.Metric{Name: "host_fs_avail_bytes", Value: float64(avail), Timestamp: now, Labels: labels},
			collector.Metric{Name: "host_fs_used_pct", Value: usedPct, Timestamp: now, Labels: labels},
		)
	}

	return metrics, nil
}

type mountEntry struct {
	device     string
	mountpoint string
	fstype     string
}

func parseMounts() ([]mountEntry, error) {
	// Try /proc/mounts first (Linux), fall back to /etc/mtab
	path := "/proc/mounts"
	f, err := os.Open(path)
	if err != nil {
		f, err = os.Open("/etc/mtab")
		if err != nil {
			return nil, fmt.Errorf("cannot read mount info: %w", err)
		}
	}
	defer f.Close()

	var mounts []mountEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mountpoint := fields[1]
		fstype := fields[2]

		// Only real filesystems
		if !strings.HasPrefix(device, "/dev/") {
			continue
		}
		if fstype == "squashfs" || fstype == "tmpfs" || fstype == "devtmpfs" {
			continue
		}

		mounts = append(mounts, mountEntry{device: device, mountpoint: mountpoint, fstype: fstype})
	}

	return mounts, nil
}
