package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/tjst-t/clabnoc/internal/topology"
)

// handleGetTopology handles GET /api/v1/projects/{name}/topology
func (s *Server) handleGetTopology(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	topo, err := s.loadTopology(r, name)
	if err != nil {
		slog.Error("load topology", "project", name, "err", err)
		http.Error(w, fmt.Sprintf("failed to load topology: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, topo)
}

// loadTopology loads and enriches the topology for a project
func (s *Server) loadTopology(r *http.Request, name string) (*topology.Topology, error) {
	// Get labdir
	labDir, err := s.discoverer.GetProjectLabDir(r.Context(), name)
	if err != nil {
		return nil, fmt.Errorf("get labdir for project %s: %w", name, err)
	}

	// Load topology-data.json
	topoPath := filepath.Join(labDir, "topology-data.json")
	topo, err := topology.ParseFile(topoPath)
	if err != nil {
		return nil, fmt.Errorf("parse topology file %s: %w", topoPath, err)
	}

	// Enrich nodes with Docker status
	containers, err := s.discoverer.GetContainersByProject(r.Context(), name)
	if err != nil {
		slog.Warn("failed to get containers for enrichment, continuing without Docker status",
			"project", name, "err", err)
	} else {
		// Build a map from clab-node-name -> container info
		type containerInfo struct {
			id     string
			status string
			state  string
		}
		containerMap := map[string]containerInfo{}
		for _, c := range containers {
			nodeName := c.Labels["clab-node-name"]
			if nodeName == "" {
				continue
			}
			containerMap[nodeName] = containerInfo{
				id:     c.ID,
				status: c.Status,
				state:  string(c.State),
			}
		}

		// Enrich each node
		for i, node := range topo.Nodes {
			if info, ok := containerMap[node.Name]; ok {
				topo.Nodes[i].ContainerID = info.id
				topo.Nodes[i].Status = info.state
			}
		}
	}

	return topo, nil
}
