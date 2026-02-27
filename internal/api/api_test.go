package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"

	"github.com/tjst-t/clabnoc/internal/docker"
	"github.com/tjst-t/clabnoc/internal/network"
)

// mockDockerClient is a test-local mock to avoid import cycles.
type mockDockerClient struct {
	containers    []container.Summary
	inspectResult container.InspectResponse
	startErr      error
	stopErr       error
	restartErr    error
}

func (m *mockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return m.containers, nil
}

func (m *mockDockerClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return m.inspectResult, nil
}

func (m *mockDockerClient) ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	return container.ExecCreateResponse{ID: "exec-123"}, nil
}

func (m *mockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error) {
	return types.HijackedResponse{}, nil
}

func (m *mockDockerClient) ContainerExecResize(ctx context.Context, execID string, options container.ResizeOptions) error {
	return nil
}

func (m *mockDockerClient) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(strings.NewReader("")), container.PathStat{}, nil
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return m.startErr
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return m.stopErr
}

func (m *mockDockerClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	return m.restartErr
}

func (m *mockDockerClient) Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error) {
	return make(chan events.Message), make(chan error)
}

// mockFaultOp implements network.FaultOperator for API tests.
type mockFaultOp struct{}

func (m *mockFaultOp) LinkSetDown(ctx context.Context, containerID, ifName string) error { return nil }
func (m *mockFaultOp) LinkSetUp(ctx context.Context, containerID, ifName string) error   { return nil }
func (m *mockFaultOp) ApplyNetem(ctx context.Context, containerID, ifName string, params *network.NetemParams) error {
	return nil
}
func (m *mockFaultOp) ClearNetem(ctx context.Context, containerID, ifName string) error { return nil }

func setupTestServer(t *testing.T) (*Server, *mockDockerClient) {
	t.Helper()

	// Create temporary topology data file
	labDir := t.TempDir()
	topoData, err := os.ReadFile("../../testdata/topology-v073.json")
	if err != nil {
		t.Fatalf("failed to read test topology: %v", err)
	}
	if err := os.WriteFile(filepath.Join(labDir, "topology-data.json"), topoData, 0644); err != nil {
		t.Fatalf("failed to write topology: %v", err)
	}

	mock := &mockDockerClient{
		containers: []container.Summary{
			{
				ID:    "abc123",
				State: "running",
				Labels: map[string]string{
					"containerlab":      "dc-fabric",
					"clab-node-name":    "spine1",
					"clab-node-lab-dir": filepath.Join(labDir, "spine1"),
				},
			},
			{
				ID:    "def456",
				State: "running",
				Labels: map[string]string{
					"containerlab":      "dc-fabric",
					"clab-node-name":    "leaf1",
					"clab-node-lab-dir": filepath.Join(labDir, "leaf1"),
				},
			},
			{
				ID:    "ghi789",
				State: "running",
				Labels: map[string]string{
					"containerlab":      "dc-fabric",
					"clab-node-name":    "server1",
					"clab-node-lab-dir": filepath.Join(labDir, "server1"),
				},
			},
		},
	}

	fm := network.NewFaultManager(&mockFaultOp{})

	server := &Server{
		Docker:       mock,
		FaultManager: fm,
	}

	return server, mock
}

func TestListProjects(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var projects []docker.ProjectInfo
	if err := json.NewDecoder(rec.Body).Decode(&projects); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "dc-fabric" {
		t.Errorf("expected project name dc-fabric, got %s", projects[0].Name)
	}
}

func TestGetTopology(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	req := httptest.NewRequest("GET", "/api/v1/projects/dc-fabric/topology", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["name"] != "dc-fabric" {
		t.Errorf("expected name dc-fabric, got %v", result["name"])
	}
}

func TestListNodes(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	req := httptest.NewRequest("GET", "/api/v1/projects/dc-fabric/nodes", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var nodes []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&nodes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(nodes))
	}
}

func TestGetNode(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	req := httptest.NewRequest("GET", "/api/v1/projects/dc-fabric/nodes/spine1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var node map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&node); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if node["name"] != "spine1" {
		t.Errorf("expected name spine1, got %v", node["name"])
	}
}

func TestGetNodeNotFound(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	req := httptest.NewRequest("GET", "/api/v1/projects/dc-fabric/nodes/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestNodeAction(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	body := strings.NewReader(`{"action": "stop"}`)
	req := httptest.NewRequest("POST", "/api/v1/projects/dc-fabric/nodes/spine1/action", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNodeActionInvalid(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	body := strings.NewReader(`{"action": "invalid"}`)
	req := httptest.NewRequest("POST", "/api/v1/projects/dc-fabric/nodes/spine1/action", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestListLinks(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	req := httptest.NewRequest("GET", "/api/v1/projects/dc-fabric/links", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var links []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&links); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestTopologyIncludesFaultState(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	// Pre-set fault state via FaultManager
	linkID := "spine1:e1-1__leaf1:e1-49"
	server.FaultManager.SetEndpointMapping(linkID,
		&network.EndpointTarget{ContainerID: "abc123", Interface: "e1-1"},
		&network.EndpointTarget{ContainerID: "def456", Interface: "e1-49"},
	)
	if err := server.FaultManager.LinkDown(context.Background(), linkID); err != nil {
		t.Fatalf("failed to set link down: %v", err)
	}

	// Fetch topology and verify link state is "down"
	req := httptest.NewRequest("GET", "/api/v1/projects/dc-fabric/topology", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var topo struct {
		Links []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"links"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&topo); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	for _, l := range topo.Links {
		if l.ID == linkID {
			if l.State != "down" {
				t.Errorf("expected link %s state 'down', got '%s'", linkID, l.State)
			}
			return
		}
	}
	t.Errorf("link %s not found in topology response", linkID)
}

func TestFaultInjectionAutoResolve(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	body := strings.NewReader(`{"action": "down"}`)
	req := httptest.NewRequest("POST", "/api/v1/projects/dc-fabric/links/spine1:e1-1__leaf1:e1-49/fault", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Auto-resolve now uses FindContainerByNode (no /proc access needed),
	// and mockFaultOp is a no-op, so this should succeed.
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
