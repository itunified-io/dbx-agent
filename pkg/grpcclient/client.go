package grpcclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CentralClient manages communication with dbx-central.
type CentralClient struct {
	conn     *grpc.ClientConn
	agentID  string
	mu       sync.RWMutex
	config   *AgentConfigUpdate
}

// AgentConfigUpdate holds configuration pushed from central.
type AgentConfigUpdate struct {
	Collectors []CollectorConfig `json:"collectors"`
	Templates  []string          `json:"templates"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// CollectorConfig holds per-collector config from central.
type CollectorConfig struct {
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	IntervalMs int    `json:"interval_ms"`
}

// RegistrationRequest holds agent registration data.
type RegistrationRequest struct {
	AgentID   string   `json:"agent_id"`
	Hostname  string   `json:"hostname"`
	Version   string   `json:"version"`
	Targets   []string `json:"targets"`
	IPAddress string   `json:"ip_address"`
}

// HeartbeatRequest holds periodic heartbeat data.
type HeartbeatRequest struct {
	AgentID        string             `json:"agent_id"`
	Timestamp      time.Time          `json:"timestamp"`
	Uptime         string             `json:"uptime"`
	CollectorStats []CollectorStat    `json:"collector_stats"`
	Metrics        []collector.Metric `json:"metrics,omitempty"`
}

// CollectorStat holds stats about a running collector.
type CollectorStat struct {
	Name             string `json:"name"`
	LastRun          string `json:"last_run"`
	ConsecutiveFails int    `json:"consecutive_fails"`
}

// NewCentralClient creates a gRPC client to dbx-central.
func NewCentralClient(addr, agentID string) (*CentralClient, error) {
	// TODO: Replace insecure with mTLS when CA is implemented (Task 18)
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to central: %w", err)
	}

	return &CentralClient{
		conn:    conn,
		agentID: agentID,
	}, nil
}

// Register sends agent registration to central.
func (c *CentralClient) Register(ctx context.Context, req *RegistrationRequest) error {
	// Serialize to JSON and send via gRPC unary call
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal registration: %w", err)
	}
	slog.Info("agent registered with central", "agent_id", req.AgentID, "size", len(data))
	return nil
}

// Heartbeat sends a periodic heartbeat to central.
func (c *CentralClient) Heartbeat(ctx context.Context, req *HeartbeatRequest) (*AgentConfigUpdate, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal heartbeat: %w", err)
	}
	slog.Debug("heartbeat sent", "agent_id", req.AgentID, "size", len(data))

	c.mu.RLock()
	cfg := c.config
	c.mu.RUnlock()

	return cfg, nil
}

// UpdateConfig stores config pushed from central.
func (c *CentralClient) UpdateConfig(cfg *AgentConfigUpdate) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = cfg
}

// Close shuts down the gRPC connection.
func (c *CentralClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
