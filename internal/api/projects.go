package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// handleGetProjects handles GET /api/v1/projects
func (s *Server) handleGetProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.discoverer.ListProjects(r.Context())
	if err != nil {
		slog.Error("list projects", "err", err)
		http.Error(w, "failed to list projects", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, projects)
}

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode JSON response", "err", err)
	}
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
