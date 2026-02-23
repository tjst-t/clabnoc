package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/gorilla/websocket"
)

// Event represents an event sent to WebSocket clients.
type Event struct {
	Type    string      `json:"type"`
	Project string      `json:"project,omitempty"`
	Data    interface{} `json:"data"`
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	projectFilter := r.URL.Query().Get("project")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed for events", "error", err)
		return
	}
	defer ws.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Monitor for client disconnect
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	f := filters.NewArgs()
	f.Add("label", "containerlab")
	f.Add("type", "container")

	msgCh, errCh := s.Docker.Events(ctx, events.ListOptions{Filters: f})

	// Send initial ping
	writeEvent(ws, Event{Type: "connected", Data: map[string]string{"status": "ok"}})

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			if err != nil {
				slog.Error("docker events error", "error", err)
				return
			}
		case msg := <-msgCh:
			project := msg.Actor.Attributes["containerlab"]
			if projectFilter != "" && project != projectFilter {
				continue
			}

			evt := convertDockerEvent(msg, project)
			if evt != nil {
				writeEvent(ws, *evt)
			}
		}
	}
}

func convertDockerEvent(msg events.Message, project string) *Event {
	nodeName := msg.Actor.Attributes["clab-node-name"]

	switch msg.Action {
	case "start":
		return &Event{
			Type:    "node_status_changed",
			Project: project,
			Data: map[string]string{
				"node":   nodeName,
				"status": "running",
			},
		}
	case "die", "stop":
		return &Event{
			Type:    "node_status_changed",
			Project: project,
			Data: map[string]string{
				"node":   nodeName,
				"status": "stopped",
			},
		}
	case "create":
		return &Event{
			Type: "project_changed",
			Data: map[string]string{
				"action":  "updated",
				"project": project,
			},
		}
	case "destroy":
		return &Event{
			Type: "project_changed",
			Data: map[string]string{
				"action":  "updated",
				"project": project,
			},
		}
	default:
		return nil
	}
}

func writeEvent(ws *websocket.Conn, evt Event) {
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Error("failed to marshal event", "error", err)
		return
	}
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		slog.Debug("failed to write event to websocket", "error", err)
	}
}
