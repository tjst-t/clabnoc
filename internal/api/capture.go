package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/tjst-t/clabnoc/internal/capture"
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

// captureStream handles WebSocket connections for live packet streaming.
func (s *Server) captureStream(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	linkID, _ := url.PathUnescape(chi.URLParam(r, "id"))

	if s.StreamExecutor == nil || s.VethResolver == nil {
		http.Error(w, "capture streaming not available", http.StatusServiceUnavailable)
		return
	}

	// Resolve host veth
	hostVeth, err := s.resolveHostVeth(r.Context(), name, linkID)
	if err != nil {
		slog.Error("failed to resolve host veth for stream", "link", linkID, "error", err)
		http.Error(w, "failed to resolve interface: "+err.Error(), http.StatusInternalServerError)
		return
	}

	bpfFilter := r.URL.Query().Get("bpf_filter")

	// Upgrade to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer ws.Close()

	// Start tcpdump stream
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	stdout, cmd, err := s.StreamExecutor.StartStream(ctx, hostVeth, bpfFilter)
	if err != nil {
		slog.Error("failed to start capture stream", "link", linkID, "error", err)
		ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","data":"failed to start stream"}`))
		return
	}

	// Client control message reader
	var paused bool
	var mu sync.Mutex

	// Read control messages from client
	go func() {
		defer cancel()
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}
			var ctrl struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(msg, &ctrl) == nil {
				mu.Lock()
				switch ctrl.Type {
				case "pause":
					paused = true
				case "resume":
					paused = false
				}
				mu.Unlock()
			}
		}
	}()

	// Read stdout line by line and send as JSON
	scanner := bufio.NewScanner(stdout)
	seqNo := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break
		default:
		}

		mu.Lock()
		isPaused := paused
		mu.Unlock()

		if isPaused {
			continue
		}

		line := scanner.Text()
		seqNo++

		pkt, err := capture.ParseTcpdumpLine(line, seqNo)
		if err != nil {
			continue // skip unparseable lines (e.g., tcpdump startup messages)
		}

		data, err := json.Marshal(pkt)
		if err != nil {
			continue
		}

		ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}

	// Clean up: kill tcpdump
	if cmd.Process != nil {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
}
