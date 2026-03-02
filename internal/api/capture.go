package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/topology"
)

type captureRequest struct {
	Action    string `json:"action"` // "start" or "stop"
	BPFFilter string `json:"bpf_filter,omitempty"`
}

func (s *Server) captureAction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	linkID, _ := url.PathUnescape(chi.URLParam(r, "id"))

	if s.CaptureManager == nil || s.VethResolver == nil {
		http.Error(w, "capture not available", http.StatusServiceUnavailable)
		return
	}

	var req captureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	switch req.Action {
	case "start":
		// Resolve link -> container -> host veth
		hostVeth, err := s.resolveHostVeth(r.Context(), name, linkID)
		if err != nil {
			slog.Error("failed to resolve host veth", "link", linkID, "error", err)
			http.Error(w, "failed to resolve interface: "+err.Error(), http.StatusInternalServerError)
			return
		}

		session, err := s.CaptureManager.Start(r.Context(), linkID, hostVeth, req.BPFFilter)
		if err != nil {
			slog.Error("failed to start capture", "link", linkID, "error", err)
			http.Error(w, "failed to start capture: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)

	case "stop":
		if err := s.CaptureManager.Stop(linkID); err != nil {
			slog.Error("failed to stop capture", "link", linkID, "error", err)
			http.Error(w, "failed to stop capture: "+err.Error(), http.StatusInternalServerError)
			return
		}

		session := s.CaptureManager.GetSession(linkID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)

	default:
		http.Error(w, "invalid action: use 'start' or 'stop'", http.StatusBadRequest)
	}
}

func (s *Server) captureDownload(w http.ResponseWriter, r *http.Request) {
	linkID, _ := url.PathUnescape(chi.URLParam(r, "id"))

	if s.CaptureManager == nil {
		http.Error(w, "capture not available", http.StatusServiceUnavailable)
		return
	}

	filePath, err := s.CaptureManager.GetFilePath(linkID)
	if err != nil {
		http.Error(w, "no capture file available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.tcpdump.pcap")
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	http.ServeFile(w, r, filePath)
}

// resolveHostVeth resolves a link's A-side endpoint to the host-side veth interface name.
func (s *Server) resolveHostVeth(ctx context.Context, projectName, linkID string) (string, error) {
	topo, err := docker.GetProjectTopology(ctx, s.Docker, projectName)
	if err != nil {
		return "", err
	}

	var targetLink *topology.Link
	for _, l := range topo.Links {
		if l.ID == linkID {
			targetLink = &l
			break
		}
	}
	if targetLink == nil {
		return "", fmt.Errorf("link %q not found", linkID)
	}

	// Use the A-side endpoint for capture
	ctr, err := docker.FindContainerByNode(ctx, s.Docker, projectName, targetLink.A.Node)
	if err != nil {
		return "", err
	}

	hostVeth, err := s.VethResolver.Resolve(ctx, ctr.ID, targetLink.A.Interface)
	if err != nil {
		return "", err
	}

	return hostVeth, nil
}
