package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// LogFileSink writes metrics to a local JSON log file (for debugging/offline mode).
type LogFileSink struct {
	path string
	mu   sync.Mutex
	file *os.File
}

// NewLogFileSink creates a log file sink.
func NewLogFileSink(path string) *LogFileSink {
	return &LogFileSink{path: path}
}

func (s *LogFileSink) Name() string { return "logfile" }

func (s *LogFileSink) Send(ctx context.Context, metrics []collector.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file == nil {
		f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
		s.file = f
	}

	for _, m := range metrics {
		line, err := json.Marshal(m)
		if err != nil {
			slog.Warn("failed to marshal metric", "name", m.Name, "error", err)
			continue
		}
		s.file.Write(line)
		s.file.Write([]byte("\n"))
	}

	return nil
}

func (s *LogFileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
