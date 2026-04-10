package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// ConfigStoreSink sends metrics to the dbx-central config store via HTTP POST.
type ConfigStoreSink struct {
	url    string
	client *http.Client
}

// NewConfigStoreSink creates a config store HTTP sink.
func NewConfigStoreSink(url string) *ConfigStoreSink {
	return &ConfigStoreSink{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *ConfigStoreSink) Name() string { return "config_store" }

func (s *ConfigStoreSink) Send(ctx context.Context, metrics []collector.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("config_store returned %d", resp.StatusCode)
	}

	return nil
}

func (s *ConfigStoreSink) Close() error {
	s.client.CloseIdleConnections()
	return nil
}
