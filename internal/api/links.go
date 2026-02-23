package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/network"
	"github.com/tjst-t/clabnoc/internal/topology"
)

type linkResponse struct {
	topology.Link
	HostVethA string `json:"host_veth_a,omitempty"`
	HostVethZ string `json:"host_veth_z,omitempty"`
}

func (s *Server) listLinks(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology for links", "project", name, "error", err)
		http.Error(w, "failed to get topology", http.StatusInternalServerError)
		return
	}

	links := make([]linkResponse, len(topo.Links))
	for i, l := range topo.Links {
		lr := linkResponse{Link: l}
		if s.FaultManager != nil {
			state := s.FaultManager.GetState(l.ID)
			lr.State = state.State
			lr.Netem = state.Netem
			lr.HostVethA = state.HostVethA
			lr.HostVethZ = state.HostVethZ
		}
		links[i] = lr
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func (s *Server) getLink(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	linkID, _ := url.PathUnescape(chi.URLParam(r, "id"))

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology for link", "project", name, "error", err)
		http.Error(w, "failed to get topology", http.StatusInternalServerError)
		return
	}

	for _, l := range topo.Links {
		if l.ID == linkID {
			lr := linkResponse{Link: l}
			if s.FaultManager != nil {
				state := s.FaultManager.GetState(l.ID)
				lr.State = state.State
				lr.Netem = state.Netem
				lr.HostVethA = state.HostVethA
				lr.HostVethZ = state.HostVethZ
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(lr)
			return
		}
	}

	http.Error(w, "link not found", http.StatusNotFound)
}

type faultRequest struct {
	Action string          `json:"action"`
	Netem  *network.NetemParams `json:"netem,omitempty"`
}

func (s *Server) injectFault(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	linkID, _ := url.PathUnescape(chi.URLParam(r, "id"))

	if s.FaultManager == nil {
		http.Error(w, "fault injection not available", http.StatusServiceUnavailable)
		return
	}

	var req faultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Find the link first to validate
	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		http.Error(w, "failed to get topology", http.StatusInternalServerError)
		return
	}

	var found bool
	for _, l := range topo.Links {
		if l.ID == linkID {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}

	switch req.Action {
	case "down":
		err = s.FaultManager.LinkDown(linkID)
	case "up":
		err = s.FaultManager.LinkUp(linkID)
	case "netem":
		if req.Netem == nil {
			http.Error(w, "netem parameters required", http.StatusBadRequest)
			return
		}
		err = s.FaultManager.ApplyNetem(linkID, req.Netem)
	case "clear_netem":
		err = s.FaultManager.ClearNetem(linkID)
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	if err != nil {
		slog.Error("fault injection failed", "link", linkID, "action", req.Action, "error", err)
		http.Error(w, "fault injection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
