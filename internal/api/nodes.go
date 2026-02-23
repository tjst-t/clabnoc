package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/go-chi/chi/v5"
	"github.com/tjst-t/clabnoc/internal/docker"
)

func (s *Server) listNodes(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology for nodes", "project", name, "error", err)
		http.Error(w, "failed to get topology", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topo.Nodes)
}

func (s *Server) getNode(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology for node", "project", name, "error", err)
		http.Error(w, "failed to get topology", http.StatusInternalServerError)
		return
	}

	for _, node := range topo.Nodes {
		if node.Name == nodeName {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(node)
			return
		}
	}

	http.Error(w, "node not found", http.StatusNotFound)
}

type actionRequest struct {
	Action string `json:"action"`
}

func (s *Server) nodeAction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")

	var req actionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, err := docker.FindContainerByNode(r.Context(), s.Docker, name, nodeName)
	if err != nil {
		slog.Error("failed to find container", "project", name, "node", nodeName, "error", err)
		http.Error(w, "container not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	switch req.Action {
	case "start":
		err = s.Docker.ContainerStart(ctx, c.ID, container.StartOptions{})
	case "stop":
		err = s.Docker.ContainerStop(ctx, c.ID, container.StopOptions{})
	case "restart":
		err = s.Docker.ContainerRestart(ctx, c.ID, container.StopOptions{})
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	if err != nil {
		slog.Error("failed to perform action", "action", req.Action, "node", nodeName, "error", err)
		http.Error(w, "action failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
