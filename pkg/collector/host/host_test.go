package host_test

import (
	"testing"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector/host"
	"github.com/stretchr/testify/assert"
)

// Unit tests for collector metadata (work on all platforms)

func TestCPUCollector_Meta(t *testing.T) {
	c := host.NewCPUCollector(15000)
	assert.Equal(t, "host.cpu", c.Name())
	assert.Equal(t, 15*time.Second, c.Interval())
	assert.True(t, c.Enabled())
}

func TestMemoryCollector_Meta(t *testing.T) {
	c := host.NewMemoryCollector(30000)
	assert.Equal(t, "host.memory", c.Name())
	assert.Equal(t, 30*time.Second, c.Interval())
	assert.True(t, c.Enabled())
}

func TestDiskCollector_Meta(t *testing.T) {
	c := host.NewDiskCollector(15000)
	assert.Equal(t, "host.disk", c.Name())
	assert.Equal(t, 15*time.Second, c.Interval())
	assert.True(t, c.Enabled())
}

func TestNetworkCollector_Meta(t *testing.T) {
	c := host.NewNetworkCollector(15000)
	assert.Equal(t, "host.network", c.Name())
	assert.Equal(t, 15*time.Second, c.Interval())
	assert.True(t, c.Enabled())
}

func TestProcessCollector_Meta(t *testing.T) {
	c := host.NewProcessCollector(30000)
	assert.Equal(t, "host.process", c.Name())
	assert.Equal(t, 30*time.Second, c.Interval())
	assert.True(t, c.Enabled())
}

func TestFilesystemCollector_Meta(t *testing.T) {
	c := host.NewFilesystemCollector(60000)
	assert.Equal(t, "host.filesystem", c.Name())
	assert.Equal(t, time.Minute, c.Interval())
	assert.True(t, c.Enabled())
}
