package grpcclient_test

import (
	"context"
	"testing"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/grpcclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCentralClient_Register(t *testing.T) {
	// Use a dummy address - we test serialization, not actual connection
	client, err := grpcclient.NewCentralClient("localhost:0", "test-agent")
	require.NoError(t, err)
	defer client.Close()

	req := &grpcclient.RegistrationRequest{
		AgentID:  "test-agent",
		Hostname: "prod-db01",
		Version:  "1.0.0",
		Targets:  []string{"oracle-orcl", "pg-main"},
	}

	err = client.Register(context.Background(), req)
	require.NoError(t, err)
}

func TestCentralClient_Heartbeat(t *testing.T) {
	client, err := grpcclient.NewCentralClient("localhost:0", "test-agent")
	require.NoError(t, err)
	defer client.Close()

	req := &grpcclient.HeartbeatRequest{
		AgentID:   "test-agent",
		Timestamp: time.Now(),
		Uptime:    "1h30m",
	}

	cfg, err := client.Heartbeat(context.Background(), req)
	require.NoError(t, err)
	assert.Nil(t, cfg) // No config pushed yet
}

func TestCentralClient_UpdateConfig(t *testing.T) {
	client, err := grpcclient.NewCentralClient("localhost:0", "test-agent")
	require.NoError(t, err)
	defer client.Close()

	update := &grpcclient.AgentConfigUpdate{
		Collectors: []grpcclient.CollectorConfig{
			{Name: "oracle.sessions", Enabled: true, IntervalMs: 15000},
		},
		UpdatedAt: time.Now(),
	}

	client.UpdateConfig(update)

	// Verify config is returned on next heartbeat
	cfg, err := client.Heartbeat(context.Background(), &grpcclient.HeartbeatRequest{
		AgentID: "test-agent",
	})
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Len(t, cfg.Collectors, 1)
}
