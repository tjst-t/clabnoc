package api

import (
	"testing"
)

func TestCalculateCPUPercent(t *testing.T) {
	tests := []struct {
		name     string
		stats    dockerStatsJSON
		wantMin  float64
		wantMax  float64
	}{
		{
			name: "normal usage",
			stats: dockerStatsJSON{
				CPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 2000},
					SystemCPUUsage: 20000,
					OnlineCPUs:     4,
				},
				PreCPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 1000},
					SystemCPUUsage: 10000,
					OnlineCPUs:     4,
				},
			},
			wantMin: 39.0,
			wantMax: 41.0,
		},
		{
			name: "zero system delta",
			stats: dockerStatsJSON{
				CPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 1000},
					SystemCPUUsage: 10000,
					OnlineCPUs:     4,
				},
				PreCPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 1000},
					SystemCPUUsage: 10000,
					OnlineCPUs:     4,
				},
			},
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name: "zero online cpus defaults to 1",
			stats: dockerStatsJSON{
				CPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 2000},
					SystemCPUUsage: 20000,
					OnlineCPUs:     0,
				},
				PreCPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 1000},
					SystemCPUUsage: 10000,
					OnlineCPUs:     0,
				},
			},
			wantMin: 9.0,
			wantMax: 11.0,
		},
		{
			name: "negative cpu delta",
			stats: dockerStatsJSON{
				CPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 500},
					SystemCPUUsage: 20000,
					OnlineCPUs:     4,
				},
				PreCPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 1000},
					SystemCPUUsage: 10000,
					OnlineCPUs:     4,
				},
			},
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name: "clamped to max",
			stats: dockerStatsJSON{
				CPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 100000},
					SystemCPUUsage: 100,
					OnlineCPUs:     4,
				},
				PreCPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 0},
					SystemCPUUsage: 0,
					OnlineCPUs:     4,
				},
			},
			wantMin: 400.0,
			wantMax: 400.0,
		},
		{
			name: "single CPU",
			stats: dockerStatsJSON{
				CPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 600},
					SystemCPUUsage: 2000,
					OnlineCPUs:     1,
				},
				PreCPUStats: cpuStats{
					CPUUsage:       struct{ TotalUsage uint64 `json:"total_usage"` }{TotalUsage: 100},
					SystemCPUUsage: 1000,
					OnlineCPUs:     1,
				},
			},
			wantMin: 49.0,
			wantMax: 51.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCPUPercent(tt.stats)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateCPUPercent() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
