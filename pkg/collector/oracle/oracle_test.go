package oracle_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	oracol "github.com/itunified-io/dbx-agent/pkg/collector/oracle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Sessions ---

func TestSessionsCollector_Name(t *testing.T) {
	c := oracol.NewSessionsCollector(nil, 15000, "prod-orcl")
	assert.Equal(t, "oracle.sessions", c.Name())
}

func TestSessionsCollector_Collect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"total", "active", "blocked", "max"}).
			AddRow(100.0, 25.0, 3.0, 500.0))

	c := oracol.NewSessionsCollector(db, 15000, "prod-orcl")
	metrics, err := c.Collect(t.Context())
	require.NoError(t, err)
	assert.Len(t, metrics, 5)
	assert.Equal(t, "oracle_session_total", metrics[0].Name)
	assert.Equal(t, 100.0, metrics[0].Value)
	assert.InDelta(t, 20.0, metrics[4].Value, 0.01) // usage_pct = 100/500*100
}

// --- Tablespaces ---

func TestTablespacesCollector_Name(t *testing.T) {
	c := oracol.NewTablespacesCollector(nil, 60000, "prod-orcl")
	assert.Equal(t, "oracle.tablespaces", c.Name())
}

func TestTablespacesCollector_Collect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT tablespace_name").WillReturnRows(
		sqlmock.NewRows([]string{"tablespace_name", "used_percent"}).
			AddRow("USERS", 45.5).
			AddRow("SYSTEM", 72.3))

	c := oracol.NewTablespacesCollector(db, 60000, "prod-orcl")
	metrics, err := c.Collect(t.Context())
	require.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "oracle_tablespace_used_pct", metrics[0].Name)
	assert.Equal(t, 45.5, metrics[0].Value)
}

// --- Waits ---

func TestWaitsCollector_Name(t *testing.T) {
	c := oracol.NewWaitsCollector(nil, 15000, "prod-orcl")
	assert.Equal(t, "oracle.waits", c.Name())
}

func TestWaitsCollector_Collect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT event").WillReturnRows(
		sqlmock.NewRows([]string{"event", "wait_class", "time_waited_ms", "total_waits"}).
			AddRow("db file sequential read", "User I/O", 5000.0, int64(1000)).
			AddRow("log file sync", "Commit", 2000.0, int64(500)))

	c := oracol.NewWaitsCollector(db, 15000, "prod-orcl")
	metrics, err := c.Collect(t.Context())
	require.NoError(t, err)
	assert.Len(t, metrics, 4) // 2 events x 2 metrics each
}

// --- SGA/PGA ---

func TestSGAPGACollector_Name(t *testing.T) {
	c := oracol.NewSGAPGACollector(nil, 30000, "prod-orcl")
	assert.Equal(t, "oracle.sga_pga", c.Name())
}

func TestSGAPGACollector_Collect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT SUM").WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(4294967296.0)) // 4GB SGA
	mock.ExpectQuery("SELECT value FROM v.pgastat").WillReturnRows(
		sqlmock.NewRows([]string{"value"}).AddRow(1073741824.0)) // 1GB PGA
	mock.ExpectQuery("SELECT ROUND").WillReturnRows(
		sqlmock.NewRows([]string{"hit_ratio"}).AddRow(99.5))

	c := oracol.NewSGAPGACollector(db, 30000, "prod-orcl")
	metrics, err := c.Collect(t.Context())
	require.NoError(t, err)
	assert.Len(t, metrics, 3)
	assert.Equal(t, "oracle_sga_total_bytes", metrics[0].Name)
	assert.Equal(t, "oracle_buffer_cache_hit_ratio", metrics[2].Name)
}

// --- Redo/Undo ---

func TestRedoUndoCollector_Name(t *testing.T) {
	c := oracol.NewRedoUndoCollector(nil, 30000, "prod-orcl")
	assert.Equal(t, "oracle.redo_undo", c.Name())
}

func TestRedoUndoCollector_Collect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(12.0))
	mock.ExpectQuery("SELECT value FROM v.parameter").WillReturnRows(
		sqlmock.NewRows([]string{"value"}).AddRow(900.0))

	c := oracol.NewRedoUndoCollector(db, 30000, "prod-orcl")
	metrics, err := c.Collect(t.Context())
	require.NoError(t, err)
	assert.Len(t, metrics, 2)
}

// --- Data Guard (license-gated) ---

func TestDataGuardCollector_Disabled(t *testing.T) {
	lic := &mockLicense{licensed: false}
	c := oracol.NewDataGuardCollector(nil, 30000, "prod-orcl", lic)
	assert.False(t, c.Enabled())
}

func TestDataGuardCollector_Enabled(t *testing.T) {
	lic := &mockLicense{licensed: true}
	c := oracol.NewDataGuardCollector(nil, 30000, "prod-orcl", lic)
	assert.True(t, c.Enabled())
}

// --- RAC (license-gated) ---

func TestRACCollector_Disabled(t *testing.T) {
	lic := &mockLicense{licensed: false}
	c := oracol.NewRACCollector(nil, 15000, "prod-orcl", lic)
	assert.False(t, c.Enabled())
}

// --- ASM (license-gated) ---

func TestASMCollector_Disabled(t *testing.T) {
	lic := &mockLicense{licensed: false}
	c := oracol.NewASMCollector(nil, 15000, "prod-orcl", lic)
	assert.False(t, c.Enabled())
}

// --- GoldenGate (license-gated) ---

func TestGoldenGateCollector_Disabled(t *testing.T) {
	lic := &mockLicense{licensed: false}
	c := oracol.NewGoldenGateCollector(nil, 30000, "prod-orcl", lic)
	assert.False(t, c.Enabled())
}

// --- AlertLog ---

func TestAlertLogCollector_Name(t *testing.T) {
	c := oracol.NewAlertLogCollector("/nonexistent", 60000, "prod-orcl")
	assert.Equal(t, "oracle.alertlog", c.Name())
	assert.Equal(t, time.Minute, c.Interval())
}

func TestAlertLogCollector_Collect(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alert.log")
	os.WriteFile(path, []byte(`
2026-04-10T10:00:00 Thread 1 advanced to log sequence 1234
ORA-00600: internal error code, arguments: [kghfre1]
2026-04-10T10:01:00 Completed checkpoint
ORA-01555: snapshot too old
ORA-12154: TNS:could not resolve the connect identifier
`), 0644)

	c := oracol.NewAlertLogCollector(path, 60000, "prod-orcl")
	metrics, err := c.Collect(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// Find total ORA errors metric
	for _, m := range metrics {
		if m.Name == "oracle_alertlog_ora_errors" {
			assert.Equal(t, float64(3), m.Value)
		}
	}
}

// --- Helpers ---

type mockLicense struct {
	licensed bool
}

func (m *mockLicense) IsLicensed(string) bool { return m.licensed }
