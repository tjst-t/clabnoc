package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// ContainerStats holds CPU and memory stats for a single container.
type ContainerStats struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryBytes uint64  `json:"memory_bytes"`
	MemoryLimit uint64  `json:"memory_limit"`
}

// StatsMessage is sent over WebSocket with all node stats.
type StatsMessage struct {
	Type  string                    `json:"type"`
	Stats map[string]ContainerStats `json:"stats"`
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed for stats", "error", err)
		return
	}
	defer ws.Close()

	ctx, cancel := context.WithCancel(r.Context())
	var once sync.Once
	safeCancel := func() { once.Do(func() { cancel() }) }
	defer safeCancel()

	// Monitor for client disconnect
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				safeCancel()
				return
			}
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Send initial stats immediately
	s.sendStats(ctx, ws, projectName)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendStats(ctx, ws, projectName)
		}
	}
}

// listProjectContainers returns running containers for a specific clab project
// using a targeted Docker filter instead of discovering all projects.
func (s *Server) listProjectContainers(ctx context.Context, projectName string) ([]container.Summary, error) {
	f := filters.NewArgs()
	f.Add("label", "containerlab="+projectName)
	f.Add("status", "running")
	return s.Docker.ContainerList(ctx, container.ListOptions{Filters: f})
}

func (s *Server) sendStats(ctx context.Context, ws *websocket.Conn, projectName string) {
	containers, err := s.listProjectContainers(ctx, projectName)
	if err != nil {
		slog.Error("failed to list containers for stats", "project", projectName, "error", err)
		return
	}

	type nodeContainer struct {
		name string
		id   string
	}
	var targets []nodeContainer
	for _, c := range containers {
		nodeName := c.Labels["clab-node-name"]
		if nodeName != "" {
			targets = append(targets, nodeContainer{name: nodeName, id: c.ID})
		}
	}

	stats := make(map[string]ContainerStats, len(targets))
	for _, c := range targets {
		resp, err := s.Docker.ContainerStatsOneShot(ctx, c.id)
		if err != nil {
			slog.Debug("failed to get stats for container", "container", c.name, "error", err)
			continue
		}

		var v dockerStatsJSON
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			resp.Body.Close()
			slog.Debug("failed to decode stats", "container", c.name, "error", err)
			continue
		}
		resp.Body.Close()

		cpuPercent := calculateCPUPercent(v)
		stats[c.name] = ContainerStats{
			CPUPercent:  cpuPercent,
			MemoryBytes: v.MemoryStats.Usage,
			MemoryLimit: v.MemoryStats.Limit,
		}
	}

	msg := StatsMessage{
		Type:  "stats",
		Stats: stats,
	}

	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal stats", "error", err)
		return
	}
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		slog.Debug("failed to write stats to websocket", "error", err)
	}
}

// dockerStatsJSON mirrors the Docker stats JSON structure (partial).
type dockerStatsJSON struct {
	CPUStats    cpuStats    `json:"cpu_stats"`
	PreCPUStats cpuStats    `json:"precpu_stats"`
	MemoryStats memoryStats `json:"memory_stats"`
}

type cpuStats struct {
	CPUUsage struct {
		TotalUsage uint64 `json:"total_usage"`
	} `json:"cpu_usage"`
	SystemCPUUsage uint64 `json:"system_cpu_usage"`
	OnlineCPUs     uint32 `json:"online_cpus"`
}

type memoryStats struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

// calculateCPUPercent calculates CPU usage percentage from Docker stats.
// Returns a value clamped to [0, numCPUs*100].
func calculateCPUPercent(v dockerStatsJSON) float64 {
	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(v.CPUStats.SystemCPUUsage) - float64(v.PreCPUStats.SystemCPUUsage)

	if systemDelta > 0.0 && cpuDelta >= 0.0 {
		numCPUs := float64(v.CPUStats.OnlineCPUs)
		if numCPUs <= 0 {
			numCPUs = 1
		}
		pct := (cpuDelta / systemDelta) * numCPUs * 100.0
		return math.Min(pct, numCPUs*100.0)
	}
	return 0.0
}
