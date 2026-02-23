package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/go-chi/chi/v5"
)

// handleGetNodes handles GET /api/v1/projects/{name}/nodes
func (s *Server) handleGetNodes(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	topo, err := s.loadTopology(r, name)
	if err != nil {
		slog.Error("load topology for nodes", "project", name, "err", err)
		http.Error(w, fmt.Sprintf("failed to load topology: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, topo.Nodes)
}

// handleGetNode handles GET /api/v1/projects/{name}/nodes/{node}
func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")

	topo, err := s.loadTopology(r, name)
	if err != nil {
		slog.Error("load topology for node", "project", name, "node", nodeName, "err", err)
		http.Error(w, fmt.Sprintf("failed to load topology: %v", err), http.StatusInternalServerError)
		return
	}

	for _, node := range topo.Nodes {
		if node.Name == nodeName {
			writeJSON(w, http.StatusOK, node)
			return
		}
	}

	writeError(w, http.StatusNotFound, fmt.Sprintf("node %s not found in project %s", nodeName, name))
}

// nodeActionRequest is the request body for POST /api/v1/projects/{name}/nodes/{node}/action
type nodeActionRequest struct {
	Action string `json:"action"` // "start", "stop", "restart"
}

// handleNodeAction handles POST /api/v1/projects/{name}/nodes/{node}/action
func (s *Server) handleNodeAction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")

	var req nodeActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch req.Action {
	case "start", "stop", "restart":
		// valid
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid action: %s (must be start, stop, or restart)", req.Action))
		return
	}

	// Find container
	c, err := s.discoverer.GetContainerByNode(r.Context(), name, nodeName)
	if err != nil {
		slog.Error("get container for action", "project", name, "node", nodeName, "err", err)
		writeError(w, http.StatusNotFound, fmt.Sprintf("node %s not found in project %s", nodeName, name))
		return
	}

	switch req.Action {
	case "start":
		if err := s.docker.ContainerStart(r.Context(), c.ID, container.StartOptions{}); err != nil {
			slog.Error("start container", "id", c.ID, "err", err)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to start container: %v", err))
			return
		}
	case "stop":
		if err := s.docker.ContainerStop(r.Context(), c.ID, container.StopOptions{}); err != nil {
			slog.Error("stop container", "id", c.ID, "err", err)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to stop container: %v", err))
			return
		}
	case "restart":
		if err := s.docker.ContainerStop(r.Context(), c.ID, container.StopOptions{}); err != nil {
			slog.Error("stop container for restart", "id", c.ID, "err", err)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to stop container: %v", err))
			return
		}
		if err := s.docker.ContainerStart(r.Context(), c.ID, container.StartOptions{}); err != nil {
			slog.Error("start container for restart", "id", c.ID, "err", err)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to start container: %v", err))
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "action": req.Action, "node": nodeName})
}
