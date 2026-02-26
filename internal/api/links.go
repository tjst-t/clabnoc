package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/network"
	"github.com/tjst-t/clabnoc/internal/topology"
)

func (s *Server) listLinks(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology for links", "project", name, "error", err)
		http.Error(w, "failed to get topology", http.StatusInternalServerError)
		return
	}

	type linkResponse struct {
		topology.Link
	}

	links := make([]linkResponse, len(topo.Links))
	for i, l := range topo.Links {
		lr := linkResponse{Link: l}
		if s.FaultManager != nil {
			state := s.FaultManager.GetState(l.ID)
			lr.State = state.State
			lr.Netem = state.Netem
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

	type linkResponse struct {
		topology.Link
	}

	for _, l := range topo.Links {
		if l.ID == linkID {
			lr := linkResponse{Link: l}
			if s.FaultManager != nil {
				state := s.FaultManager.GetState(l.ID)
				lr.State = state.State
				lr.Netem = state.Netem
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(lr)
			return
		}
	}

	http.Error(w, "link not found", http.StatusNotFound)
}

type faultRequest struct {
	Action string               `json:"action"`
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

	var targetLink *topology.Link
	for _, l := range topo.Links {
		if l.ID == linkID {
			targetLink = &l
			break
		}
	}
	if targetLink == nil {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}

	// Auto-resolve endpoint mapping if not yet set
	state := s.FaultManager.GetState(linkID)
	if state.A == nil && state.Z == nil {
		s.resolveEndpointMapping(r.Context(), name, linkID, targetLink)
	}

	ctx := r.Context()
	switch req.Action {
	case "down":
		err = s.FaultManager.LinkDown(ctx, linkID)
	case "up":
		err = s.FaultManager.LinkUp(ctx, linkID)
	case "netem":
		if req.Netem == nil {
			http.Error(w, "netem parameters required", http.StatusBadRequest)
			return
		}
		err = s.FaultManager.ApplyNetem(ctx, linkID, req.Netem)
	case "clear_netem":
		err = s.FaultManager.ClearNetem(ctx, linkID)
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

// resolveEndpointMapping finds the container IDs for a link's endpoints.
func (s *Server) resolveEndpointMapping(ctx context.Context, projectName, linkID string, link *topology.Link) {
	type endpointInfo struct {
		side   string
		node   string
		ifName string
	}

	endpoints := []endpointInfo{
		{side: "a", node: link.A.Node, ifName: link.A.Interface},
		{side: "z", node: link.Z.Node, ifName: link.Z.Interface},
	}

	var a, z *network.EndpointTarget
	for _, ep := range endpoints {
		ctr, err := docker.FindContainerByNode(ctx, s.Docker, projectName, ep.node)
		if err != nil {
			slog.Warn("endpoint resolve: container not found", "node", ep.node, "error", err)
			continue
		}

		target := &network.EndpointTarget{
			ContainerID: ctr.ID,
			Interface:   ep.ifName,
		}

		slog.Info("endpoint resolved", "link", linkID, "side", ep.side, "node", ep.node, "container", ctr.ID, "interface", ep.ifName)
		if ep.side == "a" {
			a = target
		} else {
			z = target
		}
	}

	if a != nil || z != nil {
		s.FaultManager.SetEndpointMapping(linkID, a, z)
	}
}
