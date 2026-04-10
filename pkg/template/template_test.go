package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx-agent/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleTemplate = `
name: oracle-production
description: Production Oracle database monitoring
target_type: oracle
collectors:
  - name: oracle.sessions
    enabled: true
    interval_ms: 15000
  - name: oracle.tablespaces
    enabled: true
    interval_ms: 60000
thresholds:
  oracle_session_usage_pct:
    warning: 80
    critical: 95
    operator: gt
  oracle_tablespace_used_pct:
    warning: 85
    critical: 95
    operator: gt
`

func TestLoader_LoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oracle-prod.yaml")
	os.WriteFile(path, []byte(sampleTemplate), 0644)

	loader := template.NewLoader()
	err := loader.LoadFile(path)
	require.NoError(t, err)

	tmpl, ok := loader.Get("oracle-production")
	require.True(t, ok)
	assert.Equal(t, "oracle", tmpl.TargetType)
	assert.Len(t, tmpl.Collectors, 2)
	assert.Equal(t, 80.0, tmpl.Thresholds["oracle_session_usage_pct"].Warning)
}

func TestLoader_LoadDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "oracle.yaml"), []byte(sampleTemplate), 0644)
	os.WriteFile(filepath.Join(dir, "pg.yaml"), []byte(`
name: postgres-production
target_type: postgres
collectors:
  - name: postgres.stat_activity
    enabled: true
    interval_ms: 15000
thresholds:
  pg_buffer_cache_hit_ratio:
    warning: 90
    critical: 80
    operator: lt
`), 0644)

	loader := template.NewLoader()
	err := loader.LoadDir(dir)
	require.NoError(t, err)

	all := loader.All()
	assert.Len(t, all, 2)
}

func TestLoader_RejectNoName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte(`
target_type: oracle
collectors: []
`), 0644)

	loader := template.NewLoader()
	err := loader.LoadFile(path)
	assert.ErrorContains(t, err, "template missing name")
}

func TestMergeOverride(t *testing.T) {
	base := &template.MonitoringTemplate{
		Name: "base",
		Thresholds: map[string]template.Threshold{
			"metric_a": {Warning: 80, Critical: 95, Operator: "gt"},
			"metric_b": {Warning: 10, Critical: 5, Operator: "lt"},
		},
	}

	overrides := map[string]template.Threshold{
		"metric_a": {Warning: 70, Critical: 90, Operator: "gt"},
		"metric_c": {Warning: 50, Critical: 75, Operator: "gt"},
	}

	merged := template.MergeOverride(base, overrides)
	assert.Equal(t, 70.0, merged.Thresholds["metric_a"].Warning)  // overridden
	assert.Equal(t, 10.0, merged.Thresholds["metric_b"].Warning)  // kept
	assert.Equal(t, 50.0, merged.Thresholds["metric_c"].Warning)  // added
}
