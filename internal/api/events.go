package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/gorilla/websocket"
)

// Event is the JSON event sent to the frontend
type Event struct {
	Type    string          `json:"type"`
	Project string          `json:"project,omitempty"`
	Data    json.RawMessage `json:"data"`
}

// handleEventsWS handles WS /api/v1/events
// Query params: project (optional, filter by project)
func (s *Server) handleEventsWS(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("events ws upgrade", "err", err)
		return
	}
	defer conn.Close()

	ctx := r.Context()

	// Build Docker event filters
	f := filters.NewArgs()
	f.Add("type", "container")
	if project != "" {
		f.Add("label", "containerlab="+project)
	} else {
		f.Add("label", "containerlab") // any clab container
	}

	msgCh, errCh := s.docker.Events(ctx, events.ListOptions{Filters: f})

	// Send a keepalive every 30s
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			if err != nil {
				slog.Error("docker events error", "err", err)
				return
			}
		case msg := <-msgCh:
			event := translateDockerEvent(msg)
			if event == nil {
				continue
			}
			if err := conn.WriteJSON(event); err != nil {
				return
			}
		case <-ticker.C:
			// keepalive ping
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// translateDockerEvent converts a Docker event to a clabnoc Event
func translateDockerEvent(msg events.Message) *Event {
	proj := msg.Actor.Attributes["containerlab"]
	nodeName := msg.Actor.Attributes["clab-node-name"]

	switch msg.Action {
	case "start":
		data, _ := json.Marshal(map[string]string{"node": nodeName, "status": "running"})
		return &Event{Type: "node_status_changed", Project: proj, Data: data}
	case "stop", "die":
		data, _ := json.Marshal(map[string]string{"node": nodeName, "status": "stopped"})
		return &Event{Type: "node_status_changed", Project: proj, Data: data}
	case "create":
		data, _ := json.Marshal(map[string]string{"action": "created", "project": proj})
		return &Event{Type: "project_changed", Data: data}
	case "destroy":
		data, _ := json.Marshal(map[string]string{"action": "destroyed", "project": proj})
		return &Event{Type: "project_changed", Data: data}
	default:
		return nil
	}
}
