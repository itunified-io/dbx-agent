package discovery

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Target represents a discovered database or service.
type Target struct {
	Name     string `json:"name" yaml:"name"`
	Type     string `json:"type" yaml:"type"` // oracle, postgres, host
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	SID      string `json:"sid,omitempty" yaml:"sid,omitempty"`
	Database string `json:"database,omitempty" yaml:"database,omitempty"`
}

// Discoverer finds targets of a specific type.
type Discoverer interface {
	Name() string
	Discover(ctx context.Context) ([]Target, error)
}

// Engine runs all registered discoverers.
type Engine struct {
	discoverers []Discoverer
}

// NewEngine creates a discovery engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Register adds a discoverer.
func (e *Engine) Register(d Discoverer) {
	e.discoverers = append(e.discoverers, d)
}

// DiscoverAll runs all discoverers and returns combined results.
func (e *Engine) DiscoverAll(ctx context.Context) ([]Target, error) {
	var all []Target
	for _, d := range e.discoverers {
		targets, err := d.Discover(ctx)
		if err != nil {
			slog.Warn("discovery failed", "discoverer", d.Name(), "error", err)
			continue
		}
		all = append(all, targets...)
	}
	return all, nil
}

// OracleDiscoverer finds Oracle instances via oratab.
type OracleDiscoverer struct{}

func (d *OracleDiscoverer) Name() string { return "oracle" }

func (d *OracleDiscoverer) Discover(ctx context.Context) ([]Target, error) {
	paths := []string{"/etc/oratab", "/var/opt/oracle/oratab"}
	var targets []Target

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 2 {
				continue
			}
			sid := parts[0]
			oracleHome := parts[1]

			// Check if listener is running for this SID
			targets = append(targets, Target{
				Name: fmt.Sprintf("oracle-%s", sid),
				Type: "oracle",
				Host: "localhost",
				Port: 1521,
				SID:  sid,
			})
			_ = oracleHome
		}
	}

	// Also check for PDB via SQL if we have a connection
	return targets, nil
}

// PostgresDiscoverer finds PostgreSQL instances by scanning common paths and ports.
type PostgresDiscoverer struct{}

func (d *PostgresDiscoverer) Name() string { return "postgres" }

func (d *PostgresDiscoverer) Discover(ctx context.Context) ([]Target, error) {
	var targets []Target

	// Check common PostgreSQL data directories
	pgPaths := []string{
		"/var/lib/postgresql",
		"/var/lib/pgsql",
		"/usr/local/pgsql/data",
	}

	for _, base := range pgPaths {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			pgConf := filepath.Join(base, entry.Name(), "postgresql.conf")
			if _, err := os.Stat(pgConf); err == nil {
				port := 5432
				targets = append(targets, Target{
					Name:     fmt.Sprintf("postgres-%s", entry.Name()),
					Type:     "postgres",
					Host:     "localhost",
					Port:     port,
					Database: "postgres",
				})
			}
		}
	}

	return targets, nil
}

// SQLProbe tries to connect to a target to verify it's alive.
func SQLProbe(ctx context.Context, driverName, dsn string) error {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.PingContext(ctx)
}
