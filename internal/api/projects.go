package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tjst-t/clabnoc/internal/docker"
)

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := docker.DiscoverProjects(r.Context(), s.Docker)
	if err != nil {
		slog.Error("failed to discover projects", "error", err)
		http.Error(w, "failed to discover projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}
