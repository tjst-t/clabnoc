package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventsWS_NodeStatusChanged_Start(t *testing.T) {
	msgCh := make(chan events.Message, 10)
	errCh := make(chan error, 1)

	mock := &mockDockerClient{
		eventsFn: func(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
			return msgCh, errCh
		},
	}

	server := NewServer(mock)
	ts := httptest.NewServer(server)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/events?project=test"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	require.NoError(t, err)
	defer conn.Close()

	// Send a container start event
	msgCh <- events.Message{
		Action: "start",
		Actor: events.Actor{
			Attributes: map[string]string{
				"containerlab":   "test",
				"clab-node-name": "spine1",
			},
		},
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &event))
	assert.Equal(t, "node_status_changed", event["type"])
	assert.Equal(t, "test", event["project"])

	// The "data" field is a JSON object (map), not a string
	nodeData, ok := event["data"].(map[string]interface{})
	require.True(t, ok, "data field should be a JSON object")
	assert.Equal(t, "spine1", nodeData["node"])
	assert.Equal(t, "running", nodeData["status"])
}

func TestEventsWS_NodeStatusChanged_Stop(t *testing.T) {
	msgCh := make(chan events.Message, 10)
	errCh := make(chan error, 1)

	mock := &mockDockerClient{
		eventsFn: func(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
			return msgCh, errCh
		},
	}

	server := NewServer(mock)
	ts := httptest.NewServer(server)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/events?project=test"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	require.NoError(t, err)
	defer conn.Close()

	msgCh <- events.Message{
		Action: "stop",
		Actor: events.Actor{
			Attributes: map[string]string{
				"containerlab":   "test",
				"clab-node-name": "leaf1",
			},
		},
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &event))
	assert.Equal(t, "node_status_changed", event["type"])
	assert.Equal(t, "test", event["project"])
}

func TestEventsWS_NodeStatusChanged_Die(t *testing.T) {
	msgCh := make(chan events.Message, 10)
	errCh := make(chan error, 1)

	mock := &mockDockerClient{
		eventsFn: func(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
			return msgCh, errCh
		},
	}

	server := NewServer(mock)
	ts := httptest.NewServer(server)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	require.NoError(t, err)
	defer conn.Close()

	msgCh <- events.Message{
		Action: "die",
		Actor: events.Actor{
			Attributes: map[string]string{
				"containerlab":   "proj1",
				"clab-node-name": "router1",
			},
		},
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &event))
	assert.Equal(t, "node_status_changed", event["type"])
	assert.Equal(t, "proj1", event["project"])
}

func TestEventsWS_ProjectChanged_Create(t *testing.T) {
	msgCh := make(chan events.Message, 10)
	errCh := make(chan error, 1)

	mock := &mockDockerClient{
		eventsFn: func(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
			return msgCh, errCh
		},
	}

	server := NewServer(mock)
	ts := httptest.NewServer(server)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	require.NoError(t, err)
	defer conn.Close()

	msgCh <- events.Message{
		Action: "create",
		Actor: events.Actor{
			Attributes: map[string]string{"containerlab": "new-proj"},
		},
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &event))
	assert.Equal(t, "project_changed", event["type"])
}

func TestEventsWS_ProjectChanged_Destroy(t *testing.T) {
	msgCh := make(chan events.Message, 10)
	errCh := make(chan error, 1)

	mock := &mockDockerClient{
		eventsFn: func(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
			return msgCh, errCh
		},
	}

	server := NewServer(mock)
	ts := httptest.NewServer(server)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	require.NoError(t, err)
	defer conn.Close()

	msgCh <- events.Message{
		Action: "destroy",
		Actor: events.Actor{
			Attributes: map[string]string{"containerlab": "old-proj"},
		},
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &event))
	assert.Equal(t, "project_changed", event["type"])
}

func TestEventsWS_UnknownAction_Ignored(t *testing.T) {
	msgCh := make(chan events.Message, 10)
	errCh := make(chan error, 1)

	mock := &mockDockerClient{
		eventsFn: func(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
			return msgCh, errCh
		},
	}

	server := NewServer(mock)
	ts := httptest.NewServer(server)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	require.NoError(t, err)
	defer conn.Close()

	// Send an ignored event first
	msgCh <- events.Message{
		Action: "rename",
		Actor: events.Actor{
			Attributes: map[string]string{"containerlab": "proj1"},
		},
	}

	// Then send a real event
	msgCh <- events.Message{
		Action: "start",
		Actor: events.Actor{
			Attributes: map[string]string{
				"containerlab":   "proj1",
				"clab-node-name": "spine1",
			},
		},
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &event))
	// The first real event received should be node_status_changed (the rename was ignored)
	assert.Equal(t, "node_status_changed", event["type"])
}

func TestTranslateDockerEvent_Start(t *testing.T) {
	msg := events.Message{
		Action: "start",
		Actor: events.Actor{
			Attributes: map[string]string{
				"containerlab":   "myproj",
				"clab-node-name": "spine1",
			},
		},
	}
	event := translateDockerEvent(msg)
	require.NotNil(t, event)
	assert.Equal(t, "node_status_changed", event.Type)
	assert.Equal(t, "myproj", event.Project)
}

func TestTranslateDockerEvent_Stop(t *testing.T) {
	msg := events.Message{
		Action: "stop",
		Actor: events.Actor{
			Attributes: map[string]string{
				"containerlab":   "myproj",
				"clab-node-name": "leaf1",
			},
		},
	}
	event := translateDockerEvent(msg)
	require.NotNil(t, event)
	assert.Equal(t, "node_status_changed", event.Type)
}

func TestTranslateDockerEvent_Die(t *testing.T) {
	msg := events.Message{
		Action: "die",
		Actor: events.Actor{
			Attributes: map[string]string{
				"containerlab":   "myproj",
				"clab-node-name": "leaf1",
			},
		},
	}
	event := translateDockerEvent(msg)
	require.NotNil(t, event)
	assert.Equal(t, "node_status_changed", event.Type)
}

func TestTranslateDockerEvent_Create(t *testing.T) {
	msg := events.Message{
		Action: "create",
		Actor: events.Actor{
			Attributes: map[string]string{"containerlab": "newproj"},
		},
	}
	event := translateDockerEvent(msg)
	require.NotNil(t, event)
	assert.Equal(t, "project_changed", event.Type)
	assert.Empty(t, event.Project) // project_changed events don't set top-level project
}

func TestTranslateDockerEvent_Destroy(t *testing.T) {
	msg := events.Message{
		Action: "destroy",
		Actor: events.Actor{
			Attributes: map[string]string{"containerlab": "oldproj"},
		},
	}
	event := translateDockerEvent(msg)
	require.NotNil(t, event)
	assert.Equal(t, "project_changed", event.Type)
}

func TestTranslateDockerEvent_Unknown(t *testing.T) {
	msg := events.Message{
		Action: "pause",
		Actor: events.Actor{
			Attributes: map[string]string{"containerlab": "proj"},
		},
	}
	event := translateDockerEvent(msg)
	assert.Nil(t, event)
}
