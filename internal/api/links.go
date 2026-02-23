package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tjst-t/clabnoc/internal/network"
)

// linkResponse is the API response for a link.
type linkResponse struct {
	ID        string       `json:"id"`
	A         endpointJSON `json:"a"`
	Z         endpointJSON `json:"z"`
	State     string       `json:"state"` // "up", "down", "degraded"
	Netem     *netemJSON   `json:"netem,omitempty"`
	HostVethA string       `json:"host_veth_a,omitempty"`
	HostVethZ string       `json:"host_veth_z,omitempty"`
}

type endpointJSON struct {
	Node      string `json:"node"`
	Interface string `json:"interface"`
	MAC       string `json:"mac,omitempty"`
}

type netemJSON struct {
	DelayMs          int     `json:"delay_ms"`
	JitterMs         int     `json:"jitter_ms"`
	LossPercent      float64 `json:"loss_percent"`
	CorruptPercent   float64 `json:"corrupt_percent"`
	DuplicatePercent float64 `json:"duplicate_percent"`
}

// faultRequest is the request body for POST /api/v1/projects/{name}/links/{id}/fault.
type faultRequest struct {
	Action string     `json:"action"` // "down", "up", "netem", "clear_netem"
	Netem  *netemJSON `json:"netem,omitempty"`
}

// handleGetLinks handles GET /api/v1/projects/{name}/links
func (s *Server) handleGetLinks(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	topo, err := s.loadTopology(r, name)
	if err != nil {
		slog.Error("load topology for links", "project", name, "err", err)
		http.Error(w, fmt.Sprintf("failed to load topology: %v", err), http.StatusInternalServerError)
		return
	}

	links := make([]linkResponse, 0, len(topo.Links))
	for _, link := range topo.Links {
		lr := linkResponse{
			ID: link.ID,
			A: endpointJSON{
				Node:      link.A.Node,
				Interface: link.A.Interface,
				MAC:       link.A.MAC,
			},
			Z: endpointJSON{
				Node:      link.Z.Node,
				Interface: link.Z.Interface,
				MAC:       link.Z.MAC,
			},
			State: s.faultState.GetLinkState(link.ID),
		}

		// Attach stored netem + veth info from fault state
		if fs, ok := s.faultState.Get(link.ID); ok {
			lr.HostVethA = fs.VethA
			lr.HostVethZ = fs.VethZ
			if fs.Netem != nil {
				lr.Netem = &netemJSON{
					DelayMs:          fs.Netem.DelayMs,
					JitterMs:         fs.Netem.JitterMs,
					LossPercent:      fs.Netem.LossPercent,
					CorruptPercent:   fs.Netem.CorruptPercent,
					DuplicatePercent: fs.Netem.DuplicatePercent,
				}
			}
		}

		links = append(links, lr)
	}

	writeJSON(w, http.StatusOK, links)
}

// handleGetLink handles GET /api/v1/projects/{name}/links/{id}
func (s *Server) handleGetLink(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	linkID := chi.URLParam(r, "id")

	topo, err := s.loadTopology(r, name)
	if err != nil {
		slog.Error("load topology for link", "project", name, "link", linkID, "err", err)
		http.Error(w, fmt.Sprintf("failed to load topology: %v", err), http.StatusInternalServerError)
		return
	}

	for _, link := range topo.Links {
		if link.ID == linkID {
			lr := linkResponse{
				ID: link.ID,
				A: endpointJSON{
					Node:      link.A.Node,
					Interface: link.A.Interface,
					MAC:       link.A.MAC,
				},
				Z: endpointJSON{
					Node:      link.Z.Node,
					Interface: link.Z.Interface,
					MAC:       link.Z.MAC,
				},
				State: s.faultState.GetLinkState(link.ID),
			}

			if fs, ok := s.faultState.Get(link.ID); ok {
				lr.HostVethA = fs.VethA
				lr.HostVethZ = fs.VethZ
				if fs.Netem != nil {
					lr.Netem = &netemJSON{
						DelayMs:          fs.Netem.DelayMs,
						JitterMs:         fs.Netem.JitterMs,
						LossPercent:      fs.Netem.LossPercent,
						CorruptPercent:   fs.Netem.CorruptPercent,
						DuplicatePercent: fs.Netem.DuplicatePercent,
					}
				}
			}

			writeJSON(w, http.StatusOK, lr)
			return
		}
	}

	writeError(w, http.StatusNotFound, fmt.Sprintf("link %s not found in project %s", linkID, name))
}

// handleLinkFault handles POST /api/v1/projects/{name}/links/{id}/fault
func (s *Server) handleLinkFault(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	linkID := chi.URLParam(r, "id")

	var req faultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch req.Action {
	case "down", "up", "netem", "clear_netem":
		// valid
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid action: %s (must be down, up, netem, or clear_netem)", req.Action))
		return
	}

	// Validate link exists in topology
	topo, err := s.loadTopology(r, name)
	if err != nil {
		slog.Error("load topology for fault", "project", name, "link", linkID, "err", err)
		http.Error(w, fmt.Sprintf("failed to load topology: %v", err), http.StatusInternalServerError)
		return
	}

	var linkFound bool
	for _, link := range topo.Links {
		if link.ID == linkID {
			linkFound = true
			break
		}
	}
	if !linkFound {
		writeError(w, http.StatusNotFound, fmt.Sprintf("link %s not found in project %s", linkID, name))
		return
	}

	// Get current fault state for veth names if available
	existingState, _ := s.faultState.Get(linkID)

	switch req.Action {
	case "down":
		newState := &network.LinkFaultState{
			LinkID:    linkID,
			State:     "down",
			AppliedAt: time.Now(),
		}
		if existingState != nil {
			newState.VethA = existingState.VethA
			newState.VethZ = existingState.VethZ
		}

		// Attempt actual link down if veth names are known
		if newState.VethA != "" {
			if err := s.faultManager.LinkDown(newState.VethA); err != nil {
				slog.Warn("link down veth A", "veth", newState.VethA, "err", err)
			}
		}
		if newState.VethZ != "" {
			if err := s.faultManager.LinkDown(newState.VethZ); err != nil {
				slog.Warn("link down veth Z", "veth", newState.VethZ, "err", err)
			}
		}

		s.faultState.Set(linkID, newState)

	case "up":
		if existingState != nil {
			if existingState.VethA != "" {
				if err := s.faultManager.LinkUp(existingState.VethA); err != nil {
					slog.Warn("link up veth A", "veth", existingState.VethA, "err", err)
				}
			}
			if existingState.VethZ != "" {
				if err := s.faultManager.LinkUp(existingState.VethZ); err != nil {
					slog.Warn("link up veth Z", "veth", existingState.VethZ, "err", err)
				}
			}
			if err := s.faultManager.ClearNetem(existingState.VethA); err != nil {
				slog.Debug("clear netem veth A (may not have netem)", "veth", existingState.VethA, "err", err)
			}
			if err := s.faultManager.ClearNetem(existingState.VethZ); err != nil {
				slog.Debug("clear netem veth Z (may not have netem)", "veth", existingState.VethZ, "err", err)
			}
		}
		s.faultState.Delete(linkID)

	case "netem":
		if req.Netem == nil {
			writeError(w, http.StatusBadRequest, "netem parameters required for action 'netem'")
			return
		}

		params := network.NetemParams{
			DelayMs:          req.Netem.DelayMs,
			JitterMs:         req.Netem.JitterMs,
			LossPercent:      req.Netem.LossPercent,
			CorruptPercent:   req.Netem.CorruptPercent,
			DuplicatePercent: req.Netem.DuplicatePercent,
		}

		newState := &network.LinkFaultState{
			LinkID:    linkID,
			State:     "degraded",
			Netem:     &params,
			AppliedAt: time.Now(),
		}
		if existingState != nil {
			newState.VethA = existingState.VethA
			newState.VethZ = existingState.VethZ
		}

		if newState.VethA != "" {
			if err := s.faultManager.ApplyNetem(newState.VethA, params); err != nil {
				slog.Warn("apply netem veth A", "veth", newState.VethA, "err", err)
			}
		}
		if newState.VethZ != "" {
			if err := s.faultManager.ApplyNetem(newState.VethZ, params); err != nil {
				slog.Warn("apply netem veth Z", "veth", newState.VethZ, "err", err)
			}
		}

		s.faultState.Set(linkID, newState)

	case "clear_netem":
		if existingState != nil {
			if existingState.VethA != "" {
				if err := s.faultManager.ClearNetem(existingState.VethA); err != nil {
					slog.Warn("clear netem veth A", "veth", existingState.VethA, "err", err)
				}
			}
			if existingState.VethZ != "" {
				if err := s.faultManager.ClearNetem(existingState.VethZ); err != nil {
					slog.Warn("clear netem veth Z", "veth", existingState.VethZ, "err", err)
				}
			}
		}
		s.faultState.Delete(linkID)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "action": req.Action, "link": linkID})
}
