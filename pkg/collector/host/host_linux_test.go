//go:build linux

package host_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx-agent/pkg/collector/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests that require /proc (Linux only)

func TestCPUCollector_Collect_Linux(t *testing.T) {
	c := host.NewCPUCollector(15000)
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, metrics)

	names := make(map[string]bool)
	for _, m := range metrics {
		names[m.Name] = true
	}
	assert.True(t, names["host_cpu_usage_pct"])
	assert.True(t, names["host_cpu_iowait_pct"])
}

func TestMemoryCollector_Collect_Linux(t *testing.T) {
	c := host.NewMemoryCollector(15000)
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 6)

	names := make(map[string]bool)
	for _, m := range metrics {
		names[m.Name] = true
	}
	assert.True(t, names["host_memory_total_kb"])
	assert.True(t, names["host_memory_used_pct"])
}

func TestDiskCollector_Collect_Linux(t *testing.T) {
	c := host.NewDiskCollector(15000)
	// First call populates prevStats
	_, err := c.Collect(context.Background())
	require.NoError(t, err)
	// Second call produces delta metrics
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	// May or may not have metrics depending on disk devices present
	_ = metrics
}

func TestNetworkCollector_Collect_Linux(t *testing.T) {
	c := host.NewNetworkCollector(15000)
	// First call populates prevStats
	_, err := c.Collect(context.Background())
	require.NoError(t, err)
	// Second call produces delta metrics
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	_ = metrics
}

func TestProcessCollector_Collect_Linux(t *testing.T) {
	c := host.NewProcessCollector(30000)
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 4)

	names := make(map[string]bool)
	for _, m := range metrics {
		names[m.Name] = true
	}
	assert.True(t, names["host_process_total"])
	assert.True(t, names["host_process_running"])
}

func TestFilesystemCollector_Collect_Linux(t *testing.T) {
	c := host.NewFilesystemCollector(60000)
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	// Should have at least the root filesystem
	assert.NotEmpty(t, metrics)
}
