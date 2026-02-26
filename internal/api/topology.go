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

	// Merge fault injection state into links
	if s.FaultManager != nil {
		for i, l := range topo.Links {
			state := s.FaultManager.GetState(l.ID)
			topo.Links[i].State = state.State
			topo.Links[i].Netem = state.Netem
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topo)
}
