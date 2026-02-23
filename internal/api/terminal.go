package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// handleExecWS handles WS /api/v1/projects/{name}/nodes/{node}/exec
func (s *Server) handleExecWS(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	node := chi.URLParam(r, "node")

	cmd := r.URL.Query().Get("cmd")
	if cmd == "" {
		cmd = "/bin/bash"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("upgrade websocket", "err", err)
		return
	}
	defer conn.Close()

	// Find container
	c, err := s.discoverer.GetContainerByNode(r.Context(), name, node)
	if err != nil {
		slog.Error("get container for exec", "project", name, "node", node, "err", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	if err := docker.ExecWebSocket(r.Context(), s.docker, c.ID, []string{cmd}, conn); err != nil {
		slog.Debug("exec websocket closed", "project", name, "node", node, "err", err)
	}
}

// handleSSHWS handles WS /api/v1/projects/{name}/nodes/{node}/ssh
func (s *Server) handleSSHWS(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	node := chi.URLParam(r, "node")

	user := r.URL.Query().Get("user")
	if user == "" {
		user = "admin"
	}
	port := 22
	if portStr := r.URL.Query().Get("port"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	password := r.URL.Query().Get("password")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("upgrade websocket for ssh", "err", err)
		return
	}
	defer conn.Close()

	// Get node's mgmt IP from topology
	topo, err := s.loadTopology(r, name)
	if err != nil {
		slog.Error("load topology for ssh", "project", name, "node", node, "err", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: failed to load topology: "+err.Error()))
		return
	}

	var mgmtIP string
	for _, n := range topo.Nodes {
		if n.Name == node {
			mgmtIP = n.MgmtIPv4
			break
		}
	}
	if mgmtIP == "" {
		slog.Error("no mgmt IP for node", "project", name, "node", node)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: no mgmt IP for node "+node))
		return
	}

	cfg := ssh.Config{
		Host:     mgmtIP,
		Port:     port,
		User:     user,
		Password: password,
	}

	if err := ssh.ProxyWebSocket(r.Context(), cfg, conn); err != nil {
		slog.Debug("ssh websocket closed", "project", name, "node", node, "err", err)
	}
}
