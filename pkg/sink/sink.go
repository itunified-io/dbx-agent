package sink

import (
	"context"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

type Sink interface {
	Name() string
	Send(ctx context.Context, metrics []collector.Metric) error
	Close() error
}
