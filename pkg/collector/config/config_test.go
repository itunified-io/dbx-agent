package config_test

import (
	"testing"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector/config"
	"github.com/stretchr/testify/assert"
)

func TestOracleConfigCollector_Meta(t *testing.T) {
	c := config.NewOracleConfigCollector(nil, 300000, "prod-orcl")
	assert.Equal(t, "config.oracle", c.Name())
	assert.Equal(t, 5*time.Minute, c.Interval())
	assert.True(t, c.Enabled())
}

func TestPGConfigCollector_Meta(t *testing.T) {
	c := config.NewPGConfigCollector(nil, 300000, "prod-pg")
	assert.Equal(t, "config.postgres", c.Name())
	assert.True(t, c.Enabled())
}

func TestOSConfigCollector_Meta(t *testing.T) {
	c := config.NewOSConfigCollector(300000)
	assert.Equal(t, "config.os", c.Name())
	assert.True(t, c.Enabled())
}

func TestOSConfigCollector_Collect(t *testing.T) {
	c := config.NewOSConfigCollector(300000)
	metrics, err := c.Collect(t.Context())
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)
	assert.Equal(t, "config_os_num_cpu", metrics[0].Name)
}
