package collector

import (
	"context"
	"time"
)

type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

type Collector interface {
	Name() string
	Interval() time.Duration
	Enabled() bool
	Collect(ctx context.Context) ([]Metric, error)
}

type Registry struct {
	collectors []Collector
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(c Collector) {
	r.collectors = append(r.collectors, c)
}

func (r *Registry) Enabled() []Collector {
	var enabled []Collector
	for _, c := range r.collectors {
		if c.Enabled() {
			enabled = append(enabled, c)
		}
	}
	return enabled
}

func (r *Registry) All() []Collector {
	return r.collectors
}
