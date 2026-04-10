package agent_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/agent"
	"github.com/itunified-io/dbx-agent/pkg/collector"
	"github.com/stretchr/testify/assert"
)

type mockCollector struct {
	name     string
	interval time.Duration
	enabled  bool
	count    atomic.Int32
}

func (m *mockCollector) Name() string               { return m.name }
func (m *mockCollector) Interval() time.Duration     { return m.interval }
func (m *mockCollector) Enabled() bool               { return m.enabled }
func (m *mockCollector) Collect(ctx context.Context) ([]collector.Metric, error) {
	m.count.Add(1)
	return []collector.Metric{{Name: "test_metric", Value: 1.0, Timestamp: time.Now()}}, nil
}

func TestScheduler_RunsAtInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	mc := &mockCollector{name: "test", interval: 100 * time.Millisecond, enabled: true}
	reg := collector.NewRegistry()
	reg.Register(mc)

	sched := agent.NewScheduler(reg, nil)
	go sched.Run(ctx)

	<-ctx.Done()
	assert.GreaterOrEqual(t, mc.count.Load(), int32(2))
}

func TestScheduler_SkipsDisabled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	mc := &mockCollector{name: "disabled", interval: 50 * time.Millisecond, enabled: false}
	reg := collector.NewRegistry()
	reg.Register(mc)

	sched := agent.NewScheduler(reg, nil)
	go sched.Run(ctx)

	<-ctx.Done()
	assert.Equal(t, int32(0), mc.count.Load())
}
