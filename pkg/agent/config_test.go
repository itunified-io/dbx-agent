package agent_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx-agent/pkg/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.yaml")
	os.WriteFile(path, []byte(`
central:
  url: https://central.example.com:8091
  ca_fingerprint: "sha256:abc123"
agent:
  id: "agent-prod-orcl"
  host_port: 9100
tls:
  cert_file: /opt/dbx-agent/certs/agent.crt
  key_file: /opt/dbx-agent/certs/agent.key
  ca_file: /opt/dbx-agent/certs/ca.crt
vault:
  address: https://vault.example.com:8200
  auth_method: approle
sinks:
  victoria_metrics:
    url: https://vm.example.com:8428/api/v1/write
`), 0644)

	cfg, err := agent.LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, "https://central.example.com:8091", cfg.Central.URL)
	assert.Equal(t, "agent-prod-orcl", cfg.Agent.ID)
	assert.Equal(t, 9100, cfg.Agent.HostPort)
}

func TestConfigLoad_RejectHTTP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.yaml")
	os.WriteFile(path, []byte(`
central:
  url: http://central.example.com:8091
agent:
  id: "test"
  host_port: 9100
`), 0644)

	_, err := agent.LoadConfig(path)
	assert.ErrorContains(t, err, "plaintext communication not supported")
}

func TestConfigLoad_RejectMissingCentral(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.yaml")
	os.WriteFile(path, []byte(`
agent:
  id: "test"
  host_port: 9100
`), 0644)

	_, err := agent.LoadConfig(path)
	assert.ErrorContains(t, err, "central.url is required")
}
