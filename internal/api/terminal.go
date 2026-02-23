package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/tjst-t/clabnoc/internal/docker"
	sshproxy "github.com/tjst-t/clabnoc/internal/ssh"
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

func (s *Server) sshTerminal(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	nodeName := chi.URLParam(r, "node")
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "admin"
	}
	port := r.URL.Query().Get("port")
	if port == "" {
		port = "22"
	}
	password := r.URL.Query().Get("password")

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

	proxy := sshproxy.NewProxy(mgmtIP+":"+port, user, password)
	if err := proxy.Bridge(ws); err != nil {
		slog.Error("SSH proxy error", "error", err)
	}
}
