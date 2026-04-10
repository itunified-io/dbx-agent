package collector_test

import (
	"context"
	"testing"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
	"github.com/stretchr/testify/assert"
)

type testCollector struct {
	name    string
	enabled bool
}

func (t *testCollector) Name() string               { return t.name }
func (t *testCollector) Interval() time.Duration     { return time.Minute }
func (t *testCollector) Enabled() bool               { return t.enabled }
func (t *testCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	return nil, nil
}

func TestRegistry_Enabled(t *testing.T) {
	reg := collector.NewRegistry()
	reg.Register(&testCollector{name: "a", enabled: true})
	reg.Register(&testCollector{name: "b", enabled: false})
	reg.Register(&testCollector{name: "c", enabled: true})

	enabled := reg.Enabled()
	assert.Len(t, enabled, 2)
	assert.Equal(t, "a", enabled[0].Name())
	assert.Equal(t, "c", enabled[1].Name())
}

func TestRegistry_All(t *testing.T) {
	reg := collector.NewRegistry()
	reg.Register(&testCollector{name: "a", enabled: true})
	reg.Register(&testCollector{name: "b", enabled: false})

	all := reg.All()
	assert.Len(t, all, 2)
}
