package sink

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
)

// VictoriaSink sends metrics to VictoriaMetrics via the Prometheus remote write API.
type VictoriaSink struct {
	url    string
	client *http.Client
}

// NewVictoriaSink creates a VictoriaMetrics sink.
func NewVictoriaSink(url string) *VictoriaSink {
	return &VictoriaSink{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *VictoriaSink) Name() string { return "victoria_metrics" }

func (s *VictoriaSink) Send(ctx context.Context, metrics []collector.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Use VictoriaMetrics import API (line protocol)
	var buf bytes.Buffer
	for _, m := range metrics {
		line := formatPrometheusLine(m)
		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("victoria_metrics returned %d", resp.StatusCode)
	}

	return nil
}

func (s *VictoriaSink) Close() error {
	s.client.CloseIdleConnections()
	return nil
}

// formatPrometheusLine formats a metric in Prometheus exposition format.
func formatPrometheusLine(m collector.Metric) string {
	if len(m.Labels) == 0 {
		return fmt.Sprintf("%s %g %d", m.Name, m.Value, m.Timestamp.UnixMilli())
	}

	var pairs []string
	for k, v := range m.Labels {
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, v))
	}
	return fmt.Sprintf("%s{%s} %g %d", m.Name, strings.Join(pairs, ","), m.Value, m.Timestamp.UnixMilli())
}
