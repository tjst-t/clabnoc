package api

import (
	"encoding/json"
	"net/http"

	"github.com/tjst-t/clabnoc/internal/network"
)

// listBPFPresets returns the list of built-in BPF filter presets.
func (s *Server) listBPFPresets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(network.BPFPresets())
}
