package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tjst-t/clabnoc/internal/topology"
)

// setupTopologyTestDir creates a temp dir with a topology-data.json file
func setupTopologyTestDir(t *testing.T, project, topoFile string) string {
	t.Helper()
	tmpDir := t.TempDir()
	labDir := filepath.Join(tmpDir, "clab-"+project)
	err := os.MkdirAll(labDir, 0o755)
	require.NoError(t, err)

	// Copy the topology file
	srcPath := filepath.Join("..", "topology", "testdata", topoFile)
	data, err := os.ReadFile(srcPath)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(labDir, "topology-data.json"), data, 0o644)
	require.NoError(t, err)

	return labDir
}

func TestGetTopology_ReturnsTopology(t *testing.T) {
	labDir := setupTopologyTestDir(t, "test-topo", "topology-v073.json")

	// Setup mock with containers pointing to our temp dir
	mock := &mockDockerClient{
		containers: []container.Summary{
			{
				ID:    "c1",
				Names: []string{"/clab-test-topo-spine1"},
				State: "running",
				Labels: map[string]string{
					"containerlab":      "test-topo",
					"clab-node-name":    "spine1",
					"clab-node-lab-dir": labDir + "/spine1",
				},
			},
			{
				ID:    "c2",
				Names: []string{"/clab-test-topo-leaf1"},
				State: "running",
				Labels: map[string]string{
					"containerlab":      "test-topo",
					"clab-node-name":    "leaf1",
					"clab-node-lab-dir": labDir + "/leaf1",
				},
			},
		},
	}

	server := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/test-topo/topology", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var topo topology.Topology
	err := json.NewDecoder(w.Body).Decode(&topo)
	require.NoError(t, err)

	assert.Equal(t, "test-topo", topo.Name)
	assert.Len(t, topo.Nodes, 2)
	assert.Len(t, topo.Links, 1)

	// Verify nodes are sorted
	assert.Equal(t, "leaf1", topo.Nodes[0].Name)
	assert.Equal(t, "spine1", topo.Nodes[1].Name)

	// Check link
	assert.Equal(t, "spine1:e1-1__leaf1:e1-49", topo.Links[0].ID)
}

func TestGetTopology_EnrichesWithDockerStatus(t *testing.T) {
	labDir := setupTopologyTestDir(t, "test-topo", "topology-v073.json")

	mock := &mockDockerClient{
		containers: []container.Summary{
			{
				ID:    "abc123",
				Names: []string{"/clab-test-topo-spine1"},
				State: "running",
				Labels: map[string]string{
					"containerlab":      "test-topo",
					"clab-node-name":    "spine1",
					"clab-node-lab-dir": labDir + "/spine1",
				},
			},
		},
	}

	server := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/test-topo/topology", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var topo topology.Topology
	err := json.NewDecoder(w.Body).Decode(&topo)
	require.NoError(t, err)

	// Find spine1 and check it was enriched
	var spine1 *topology.Node
	for i := range topo.Nodes {
		if topo.Nodes[i].Name == "spine1" {
			spine1 = &topo.Nodes[i]
			break
		}
	}
	require.NotNil(t, spine1)
	assert.Equal(t, "abc123", spine1.ContainerID)
	assert.Equal(t, "running", spine1.Status)
}

func TestGetTopology_TopologyFileNotFound(t *testing.T) {
	// Container pointing to non-existent labdir
	mock := &mockDockerClient{
		containers: []container.Summary{
			{
				ID:    "c1",
				Names: []string{"/clab-noexist-n1"},
				State: "running",
				Labels: map[string]string{
					"containerlab":      "noexist",
					"clab-node-name":    "n1",
					"clab-node-lab-dir": "/nonexistent/path/n1",
				},
			},
		},
	}

	server := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/noexist/topology", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetTopology_WithLegacyFormat(t *testing.T) {
	labDir := setupTopologyTestDir(t, "test-legacy", "topology-legacy.json")

	mock := &mockDockerClient{
		containers: []container.Summary{
			{
				ID:    "c1",
				Names: []string{"/clab-test-legacy-spine1"},
				State: "running",
				Labels: map[string]string{
					"containerlab":      "test-legacy",
					"clab-node-name":    "spine1",
					"clab-node-lab-dir": labDir + "/spine1",
				},
			},
		},
	}

	server := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/test-legacy/topology", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var topo topology.Topology
	err := json.NewDecoder(w.Body).Decode(&topo)
	require.NoError(t, err)

	assert.Equal(t, "test-legacy", topo.Name)
	assert.Len(t, topo.Links, 1)
	assert.Equal(t, "spine1:e1-1__leaf1:e1-49", topo.Links[0].ID)
}
