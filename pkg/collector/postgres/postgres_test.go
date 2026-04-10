package postgres_test

import (
	"context"
	"testing"

	pgcol "github.com/itunified-io/dbx-agent/pkg/collector/postgres"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatActivityCollector_Name(t *testing.T) {
	c := pgcol.NewStatActivityCollector(nil, 15000, "prod-pg")
	assert.Equal(t, "postgres.stat_activity", c.Name())
}

func TestStatActivityCollector_Collect(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"state", "count"}).
		AddRow("active", int64(5)).
		AddRow("idle", int64(20)).
		AddRow("idle in transaction", int64(2))

	mock.ExpectQuery("SELECT COALESCE").WillReturnRows(rows)

	c := pgcol.NewStatActivityCollector(mock, 15000, "prod-pg")
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 3)
	assert.Equal(t, "pg_stat_activity_count", metrics[0].Name)
	assert.Equal(t, float64(5), metrics[0].Value)
}

func TestReplicationCollector_Name(t *testing.T) {
	c := pgcol.NewReplicationCollector(nil, 15000, "prod-pg")
	assert.Equal(t, "postgres.replication", c.Name())
}

func TestReplicationCollector_Collect(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"client_addr", "state", "replay_lag_bytes", "replay_lag_seconds"}).
		AddRow("10.0.0.2", "streaming", float64(1024), float64(0.5))

	mock.ExpectQuery("SELECT client_addr").WillReturnRows(rows)

	c := pgcol.NewReplicationCollector(mock, 15000, "prod-pg")
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "pg_replication_lag_bytes", metrics[0].Name)
	assert.Equal(t, float64(1024), metrics[0].Value)
}

func TestLocksCollector_Name(t *testing.T) {
	c := pgcol.NewLocksCollector(nil, 15000, "prod-pg")
	assert.Equal(t, "postgres.locks", c.Name())
}

func TestLocksCollector_Collect(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	lockRows := pgxmock.NewRows([]string{"mode", "count"}).
		AddRow("AccessShareLock", int64(10)).
		AddRow("RowExclusiveLock", int64(3))

	mock.ExpectQuery("SELECT mode").WillReturnRows(lockRows)
	mock.ExpectQuery("SELECT count").WillReturnRows(
		pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

	c := pgcol.NewLocksCollector(mock, 15000, "prod-pg")
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 3) // 2 lock modes + 1 blocked count
}

func TestBuffersCollector_Name(t *testing.T) {
	c := pgcol.NewBuffersCollector(nil, 15000, "prod-pg")
	assert.Equal(t, "postgres.buffers", c.Name())
}

func TestBuffersCollector_Collect(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT sum").WillReturnRows(
		pgxmock.NewRows([]string{"blks_hit", "blks_read"}).
			AddRow(float64(90000), float64(10000)))

	c := pgcol.NewBuffersCollector(mock, 15000, "prod-pg")
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 3)
	assert.Equal(t, "pg_buffer_cache_hit_ratio", metrics[0].Name)
	assert.InDelta(t, 90.0, metrics[0].Value, 0.01)
}

func TestStatementsCollector_Name(t *testing.T) {
	c := pgcol.NewStatementsCollector(nil, 30000, "prod-pg", 10)
	assert.Equal(t, "postgres.statements", c.Name())
}

func TestStatementsCollector_Collect(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"queryid", "calls", "total_exec_time", "mean_exec_time", "rows"}).
		AddRow("12345", int64(100), float64(500.0), float64(5.0), int64(1000))

	mock.ExpectQuery("SELECT queryid").WithArgs(10).WillReturnRows(rows)

	c := pgcol.NewStatementsCollector(mock, 30000, "prod-pg", 10)
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 4) // calls, total_time, mean_time, rows
}

func TestStatDatabaseCollector_Name(t *testing.T) {
	c := pgcol.NewStatDatabaseCollector(nil, 30000, "prod-pg")
	assert.Equal(t, "postgres.stat_database", c.Name())
}

func TestStatDatabaseCollector_Collect(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{
		"datname", "numbackends", "xact_commit", "xact_rollback",
		"tup_returned", "tup_fetched", "tup_inserted", "tup_updated", "tup_deleted", "deadlocks",
	}).AddRow("mydb", int64(10), int64(5000), int64(50),
		int64(100000), int64(80000), int64(1000), int64(500), int64(100), int64(2))

	mock.ExpectQuery("SELECT datname").WillReturnRows(rows)

	c := pgcol.NewStatDatabaseCollector(mock, 30000, "prod-pg")
	metrics, err := c.Collect(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 9) // 9 metrics per database
	assert.Equal(t, "pg_db_backends", metrics[0].Name)
}
