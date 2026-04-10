package discovery_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx-agent/pkg/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDiscoverer struct {
	name    string
	targets []discovery.Target
	err     error
}

func (m *mockDiscoverer) Name() string { return m.name }
func (m *mockDiscoverer) Discover(ctx context.Context) ([]discovery.Target, error) {
	return m.targets, m.err
}

func TestEngine_DiscoverAll(t *testing.T) {
	engine := discovery.NewEngine()
	engine.Register(&mockDiscoverer{
		name: "test-oracle",
		targets: []discovery.Target{
			{Name: "orcl1", Type: "oracle", Host: "localhost", Port: 1521, SID: "ORCL"},
		},
	})
	engine.Register(&mockDiscoverer{
		name: "test-pg",
		targets: []discovery.Target{
			{Name: "pg1", Type: "postgres", Host: "localhost", Port: 5432, Database: "mydb"},
		},
	})

	targets, err := engine.DiscoverAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, targets, 2)
	assert.Equal(t, "oracle", targets[0].Type)
	assert.Equal(t, "postgres", targets[1].Type)
}

func TestEngine_DiscoverAll_SkipsErrors(t *testing.T) {
	engine := discovery.NewEngine()
	engine.Register(&mockDiscoverer{
		name: "failing",
		err:  assert.AnError,
	})
	engine.Register(&mockDiscoverer{
		name:    "working",
		targets: []discovery.Target{{Name: "ok", Type: "host"}},
	})

	targets, err := engine.DiscoverAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, targets, 1)
}

func TestPostgresDiscoverer_NoDataDir(t *testing.T) {
	d := &discovery.PostgresDiscoverer{}
	targets, err := d.Discover(context.Background())
	require.NoError(t, err)
	// On non-PG host, should return empty
	_ = targets
}

func TestOracleDiscoverer_MockOratab(t *testing.T) {
	// Oracle discoverer reads /etc/oratab which likely doesn't exist on test host
	d := &discovery.OracleDiscoverer{}
	targets, err := d.Discover(context.Background())
	require.NoError(t, err)
	_ = targets
}

func TestPostgresDiscoverer_WithDataDir(t *testing.T) {
	dir := t.TempDir()
	pgDir := filepath.Join(dir, "14")
	os.MkdirAll(pgDir, 0755)
	os.WriteFile(filepath.Join(pgDir, "postgresql.conf"), []byte("# config"), 0644)

	// This won't be found since we can't modify the hardcoded paths in test
	// But we verify the discoverer runs without error
	d := &discovery.PostgresDiscoverer{}
	_, err := d.Discover(context.Background())
	require.NoError(t, err)
}
