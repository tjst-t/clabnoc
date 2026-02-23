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
	"github.com/vishvananda/netlink"

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

	mockOp := &mockVethOperator{}
	fm := network.NewFaultManager(mockOp)

	server := &Server{
		Docker:       mock,
		FaultManager: fm,
	}

	return server, mock
}

// mockVethOperator is a no-op veth operator for testing.
type mockVethOperator struct{}

func (m *mockVethOperator) LinkByName(name string) (netlink.Link, error) { return nil, nil }
func (m *mockVethOperator) LinkSetUp(link netlink.Link) error             { return nil }
func (m *mockVethOperator) LinkSetDown(link netlink.Link) error           { return nil }
func (m *mockVethOperator) LinkList() ([]netlink.Link, error)             { return nil, nil }
func (m *mockVethOperator) QdiscAdd(qdisc netlink.Qdisc) error           { return nil }
func (m *mockVethOperator) QdiscDel(qdisc netlink.Qdisc) error           { return nil }

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

func TestFaultInjectionNoMapping(t *testing.T) {
	server, _ := setupTestServer(t)
	router := NewRouter(server)

	body := strings.NewReader(`{"action": "down"}`)
	req := httptest.NewRequest("POST", "/api/v1/projects/dc-fabric/links/spine1:e1-1__leaf1:e1-49/fault", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Should fail because no veth mapping exists
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 (no veth mapping), got %d: %s", rec.Code, rec.Body.String())
	}
}
