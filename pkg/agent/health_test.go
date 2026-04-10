package agent_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/itunified-io/dbx-agent/pkg/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth_Healthy(t *testing.T) {
	hs := agent.NewHealthServer()
	handler := hs.Handler("test-agent", "1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp agent.HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "test-agent", resp.AgentID)
}

func TestHealth_Degraded(t *testing.T) {
	hs := agent.NewHealthServer()
	// Simulate 3 consecutive failures
	for i := 0; i < 3; i++ {
		hs.UpdateCollector("bad-collector", time.Millisecond, 0, errors.New("fail"))
	}

	handler := hs.Handler("test-agent", "1.0.0")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var resp agent.HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "degraded", resp.Status)
}

func TestHealth_Metrics(t *testing.T) {
	hs := agent.NewHealthServer()
	handler := hs.Handler("test-agent", "1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
	assert.Contains(t, w.Body.String(), "dbx_agent_uptime_seconds")
}
