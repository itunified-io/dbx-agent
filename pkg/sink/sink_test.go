package sink_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/collector"
	"github.com/itunified-io/dbx-agent/pkg/sink"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testMetrics = []collector.Metric{
	{Name: "test_metric", Value: 42.0, Timestamp: time.Now(), Labels: map[string]string{"host": "prod"}},
	{Name: "test_counter", Value: 100.0, Timestamp: time.Now(), Labels: map[string]string{}},
}

// --- VictoriaSink ---

func TestVictoriaSink_Name(t *testing.T) {
	s := sink.NewVictoriaSink("https://vm.example.com/api/v1/write")
	assert.Equal(t, "victoria_metrics", s.Name())
}

func TestVictoriaSink_Send(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := sink.NewVictoriaSink(server.URL)
	err := s.Send(context.Background(), testMetrics)
	require.NoError(t, err)
	assert.Contains(t, receivedBody, "test_metric")
	assert.Contains(t, receivedBody, `host="prod"`)
}

func TestVictoriaSink_SendEmpty(t *testing.T) {
	s := sink.NewVictoriaSink("https://vm.example.com")
	err := s.Send(context.Background(), nil)
	require.NoError(t, err)
}

// --- ConfigStoreSink ---

func TestConfigStoreSink_Name(t *testing.T) {
	s := sink.NewConfigStoreSink("https://central.example.com/api/v1/metrics")
	assert.Equal(t, "config_store", s.Name())
}

func TestConfigStoreSink_Send(t *testing.T) {
	var receivedMetrics []collector.Metric
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedMetrics)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := sink.NewConfigStoreSink(server.URL)
	err := s.Send(context.Background(), testMetrics)
	require.NoError(t, err)
	assert.Len(t, receivedMetrics, 2)
	assert.Equal(t, "test_metric", receivedMetrics[0].Name)
}

// --- LogFileSink ---

func TestLogFileSink_Name(t *testing.T) {
	s := sink.NewLogFileSink("/tmp/test.jsonl")
	assert.Equal(t, "logfile", s.Name())
}

func TestLogFileSink_Send(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.jsonl")

	s := sink.NewLogFileSink(path)
	defer s.Close()

	err := s.Send(context.Background(), testMetrics)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test_metric")
	assert.Contains(t, string(data), "test_counter")
}

func TestLogFileSink_SendEmpty(t *testing.T) {
	s := sink.NewLogFileSink("/tmp/test.jsonl")
	err := s.Send(context.Background(), nil)
	require.NoError(t, err)
}
