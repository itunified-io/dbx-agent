package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type HealthServer struct {
	startTime time.Time
	mu        sync.RWMutex
	statuses  map[string]CollectorStatus
}

type CollectorStatus struct {
	Name             string    `json:"name"`
	LastRun          time.Time `json:"last_run"`
	LastDuration     string    `json:"last_duration"`
	LastError        string    `json:"last_error,omitempty"`
	MetricCount      int       `json:"metric_count"`
	ConsecutiveFails int       `json:"consecutive_fails"`
}

type HealthResponse struct {
	Status     string                     `json:"status"`
	AgentID    string                     `json:"agent_id"`
	Version    string                     `json:"version"`
	Uptime     string                     `json:"uptime"`
	GoVersion  string                     `json:"go_version"`
	Arch       string                     `json:"arch"`
	Collectors map[string]CollectorStatus `json:"collectors"`
}

func NewHealthServer() *HealthServer {
	return &HealthServer{
		startTime: time.Now(),
		statuses:  make(map[string]CollectorStatus),
	}
}

func (h *HealthServer) UpdateCollector(name string, duration time.Duration, metricCount int, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	status := h.statuses[name]
	status.Name = name
	status.LastRun = time.Now()
	status.LastDuration = duration.String()
	status.MetricCount = metricCount
	if err != nil {
		status.LastError = err.Error()
		status.ConsecutiveFails++
	} else {
		status.LastError = ""
		status.ConsecutiveFails = 0
	}
	h.statuses[name] = status
}

func (h *HealthServer) Handler(agentID, version string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		h.mu.RLock()
		defer h.mu.RUnlock()
		status := "healthy"
		for _, cs := range h.statuses {
			if cs.ConsecutiveFails >= 3 {
				status = "degraded"
				break
			}
		}
		resp := HealthResponse{
			Status:     status,
			AgentID:    agentID,
			Version:    version,
			Uptime:     time.Since(h.startTime).String(),
			GoVersion:  runtime.Version(),
			Arch:       runtime.GOARCH,
			Collectors: h.statuses,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		h.mu.RLock()
		defer h.mu.RUnlock()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		fmt.Fprintf(w, "# HELP dbx_agent_uptime_seconds Agent uptime in seconds\n")
		fmt.Fprintf(w, "# TYPE dbx_agent_uptime_seconds gauge\n")
		fmt.Fprintf(w, "dbx_agent_uptime_seconds %f\n", time.Since(h.startTime).Seconds())
		fmt.Fprintf(w, "# HELP dbx_agent_memory_alloc_bytes Current memory allocation\n")
		fmt.Fprintf(w, "# TYPE dbx_agent_memory_alloc_bytes gauge\n")
		fmt.Fprintf(w, "dbx_agent_memory_alloc_bytes %d\n", ms.Alloc)
	})
	return mux
}
