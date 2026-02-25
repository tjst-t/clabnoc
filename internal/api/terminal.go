package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/tjst-t/clabnoc/internal/docker"
	sshproxy "github.com/tjst-t/clabnoc/internal/ssh"
	"github.com/tjst-t/clabnoc/internal/topology"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) execTerminal(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")
	cmd := r.URL.Query().Get("cmd")
	if cmd == "" {
		cmd = "/bin/bash"
	}

	c, err := docker.FindContainerByNode(r.Context(), s.Docker, name, nodeName)
	if err != nil {
		slog.Error("container not found for exec", "project", name, "node", nodeName, "error", err)
		http.Error(w, "container not found", http.StatusNotFound)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer ws.Close()

	session := docker.NewExecSession(s.Docker, c.ID, cmd)
	if err := session.Bridge(r.Context(), ws); err != nil {
		slog.Error("exec session error", "error", err)
	}
}

func (s *Server) getSSHCredentials(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")

	result, err := docker.GetProjectTopologyWithConfig(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology for SSH credentials", "project", name, "error", err)
		http.Error(w, "topology not found", http.StatusInternalServerError)
		return
	}

	var kind string
	for _, node := range result.Topology.Nodes {
		if node.Name == nodeName {
			kind = node.Kind
			break
		}
	}

	creds := topology.ResolveSSHCredentials(kind, nodeName, result.Config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
}

func (s *Server) sshTerminal(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")

	// Legacy support: if query params are provided, use them
	legacyUser := r.URL.Query().Get("user")
	legacyPort := r.URL.Query().Get("port")
	legacyPassword := r.URL.Query().Get("password")
	useLegacy := legacyUser != "" || legacyPort != ""

	topo, err := docker.GetProjectTopology(r.Context(), s.Docker, name)
	if err != nil {
		slog.Error("failed to get topology for SSH", "project", name, "error", err)
		http.Error(w, "topology not found", http.StatusInternalServerError)
		return
	}

	var mgmtIP string
	for _, node := range topo.Nodes {
		if node.Name == nodeName {
			mgmtIP = node.MgmtIPv4
			break
		}
	}
	if mgmtIP == "" {
		http.Error(w, "node has no management IP", http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer ws.Close()

	var user, password, port string

	if useLegacy {
		// Legacy query parameter mode
		user = legacyUser
		if user == "" {
			user = "admin"
		}
		port = legacyPort
		if port == "" {
			port = "22"
		}
		password = legacyPassword
	} else {
		// New mode: read credentials from first WebSocket message
		_, msg, err := ws.ReadMessage()
		if err != nil {
			slog.Error("failed to read SSH credentials from WebSocket", "error", err)
			return
		}

		var creds topology.SSHCredentials
		if err := json.Unmarshal(msg, &creds); err != nil {
			slog.Error("failed to parse SSH credentials", "error", err)
			return
		}

		user = creds.Username
		if user == "" {
			user = "admin"
		}
		password = creds.Password
		port = "22"
		if creds.Port > 0 {
			port = fmt.Sprintf("%d", creds.Port)
		}
	}

	proxy := sshproxy.NewProxy(mgmtIP+":"+port, user, password)
	if err := proxy.Bridge(ws); err != nil {
		slog.Error("SSH proxy error", "error", err)
	}
}
