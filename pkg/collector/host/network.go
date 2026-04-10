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

// NetworkCollector reads /proc/net/dev for network interface metrics.
type NetworkCollector struct {
	intervalMs int
	enabled    bool
	prevStats  map[string]netStat
}

type netStat struct {
	rxBytes   uint64
	txBytes   uint64
	rxPackets uint64
	txPackets uint64
	rxErrors  uint64
	txErrors  uint64
}

func NewNetworkCollector(intervalMs int) *NetworkCollector {
	return &NetworkCollector{intervalMs: intervalMs, enabled: true, prevStats: make(map[string]netStat)}
}

func (c *NetworkCollector) Name() string           { return "host.network" }
func (c *NetworkCollector) Interval() time.Duration { return time.Duration(c.intervalMs) * time.Millisecond }
func (c *NetworkCollector) Enabled() bool           { return c.enabled }

func (c *NetworkCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("open /proc/net/dev: %w", err)
	}
	defer f.Close()

	now := time.Now()
	var metrics []collector.Metric
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}

		rxBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		rxPackets, _ := strconv.ParseUint(fields[1], 10, 64)
		rxErrors, _ := strconv.ParseUint(fields[2], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		txPackets, _ := strconv.ParseUint(fields[9], 10, 64)
		txErrors, _ := strconv.ParseUint(fields[10], 10, 64)

		labels := map[string]string{"interface": iface}

		current := netStat{
			rxBytes: rxBytes, txBytes: txBytes,
			rxPackets: rxPackets, txPackets: txPackets,
			rxErrors: rxErrors, txErrors: txErrors,
		}

		if prev, ok := c.prevStats[iface]; ok {
			metrics = append(metrics,
				collector.Metric{Name: "host_net_rx_bytes", Value: float64(current.rxBytes - prev.rxBytes), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_net_tx_bytes", Value: float64(current.txBytes - prev.txBytes), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_net_rx_packets", Value: float64(current.rxPackets - prev.rxPackets), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_net_tx_packets", Value: float64(current.txPackets - prev.txPackets), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_net_rx_errors", Value: float64(current.rxErrors - prev.rxErrors), Timestamp: now, Labels: labels},
				collector.Metric{Name: "host_net_tx_errors", Value: float64(current.txErrors - prev.txErrors), Timestamp: now, Labels: labels},
			)
		}

		c.prevStats[iface] = current
	}

	return metrics, nil
}
