package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tjst-t/clabnoc/internal/docker"
)

func (s *Server) getTopology(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology", "project", name, "error", err)
		http.Error(w, "failed to get topology", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topo)
}
