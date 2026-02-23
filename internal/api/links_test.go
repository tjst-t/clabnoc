package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjst-t/clabnoc/internal/network"
)

// setupLinksTestServer creates a test server with topology loaded from testdata
func setupLinksTestServer(t *testing.T, project, topoFile string) (*Server, string) {
	t.Helper()
	tmpDir := t.TempDir()
	labDir := filepath.Join(tmpDir, "clab-"+project)
	err := os.MkdirAll(labDir, 0o755)
	require.NoError(t, err)

	srcPath := filepath.Join("..", "topology", "testdata", topoFile)
	data, err := os.ReadFile(srcPath)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(labDir, "topology-data.json"), data, 0o644)
	require.NoError(t, err)

	mock := &mockDockerClient{
		containers: []container.Summary{
			{
				ID:    "c1",
				Names: []string{"/clab-" + project + "-spine1"},
				State: "running",
				Labels: map[string]string{
					"containerlab":      project,
					"clab-node-name":    "spine1",
					"clab-node-lab-dir": labDir + "/spine1",
				},
			},
			{
				ID:    "c2",
				Names: []string{"/clab-" + project + "-leaf1"},
				State: "running",
				Labels: map[string]string{
					"containerlab":      project,
					"clab-node-name":    "leaf1",
					"clab-node-lab-dir": labDir + "/leaf1",
				},
			},
		},
	}

	server := NewServer(mock)
	return server, labDir
}

func TestGetLinks_ReturnsLinks(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-links", "topology-v073.json")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/test-links/links", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var links []linkResponse
	err := json.NewDecoder(w.Body).Decode(&links)
	require.NoError(t, err)

	require.Len(t, links, 1)
	assert.Equal(t, "spine1:e1-1__leaf1:e1-49", links[0].ID)
	assert.Equal(t, "spine1", links[0].A.Node)
	assert.Equal(t, "e1-1", links[0].A.Interface)
	assert.Equal(t, "leaf1", links[0].Z.Node)
	assert.Equal(t, "e1-49", links[0].Z.Interface)
}

func TestGetLinks_DefaultStateIsUp(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-links-state", "topology-v073.json")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/test-links-state/links", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var links []linkResponse
	err := json.NewDecoder(w.Body).Decode(&links)
	require.NoError(t, err)

	require.Len(t, links, 1)
	assert.Equal(t, "up", links[0].State)
	assert.Nil(t, links[0].Netem)
}

func TestGetLink_ReturnsSpecificLink(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-link-get", "topology-v073.json")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/test-link-get/links/spine1:e1-1__leaf1:e1-49", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var link linkResponse
	err := json.NewDecoder(w.Body).Decode(&link)
	require.NoError(t, err)

	assert.Equal(t, "spine1:e1-1__leaf1:e1-49", link.ID)
	assert.Equal(t, "up", link.State)
}

func TestGetLink_NotFound(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-link-notfound", "topology-v073.json")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/test-link-notfound/links/nonexistent", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLinkFault_ActionDown_UpdatesState(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-down", "topology-v073.json")
	linkID := "spine1:e1-1__leaf1:e1-49"

	body := faultRequest{Action: "down"}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/test-fault-down/links/"+linkID+"/fault",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "down", resp["action"])
	assert.Equal(t, linkID, resp["link"])
}

func TestLinkFault_ActionDown_StateReflectedInGet(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-reflect", "topology-v073.json")
	linkID := "spine1:e1-1__leaf1:e1-49"

	// Inject fault state directly
	server.faultState.Set(linkID, &network.LinkFaultState{
		LinkID: linkID,
		State:  "down",
	})

	// GET the link - should show down state
	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/test-fault-reflect/links/"+linkID, nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var link linkResponse
	err := json.NewDecoder(w.Body).Decode(&link)
	require.NoError(t, err)
	assert.Equal(t, "down", link.State)
}

func TestLinkFault_ActionUp_ClearsState(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-up", "topology-v073.json")
	linkID := "spine1:e1-1__leaf1:e1-49"

	// Set fault state directly
	server.faultState.Set(linkID, &network.LinkFaultState{
		LinkID: linkID,
		State:  "down",
	})

	// Send "up" action
	body := faultRequest{Action: "up"}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/test-fault-up/links/"+linkID+"/fault",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify fault state is cleared
	state := server.faultState.GetLinkState(linkID)
	assert.Equal(t, "up", state)
}

func TestLinkFault_ActionNetem_UpdatesState(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-netem", "topology-v073.json")
	linkID := "spine1:e1-1__leaf1:e1-49"

	body := faultRequest{
		Action: "netem",
		Netem: &netemJSON{
			DelayMs:     100,
			JitterMs:    10,
			LossPercent: 1.5,
		},
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/test-fault-netem/links/"+linkID+"/fault",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify fault state
	state := server.faultState.GetLinkState(linkID)
	assert.Equal(t, "degraded", state)

	fs, ok := server.faultState.Get(linkID)
	require.True(t, ok)
	require.NotNil(t, fs.Netem)
	assert.Equal(t, 100, fs.Netem.DelayMs)
	assert.Equal(t, 10, fs.Netem.JitterMs)
	assert.Equal(t, 1.5, fs.Netem.LossPercent)
}

func TestLinkFault_ActionNetem_ReflectedInGet(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-netem-get", "topology-v073.json")
	linkID := "spine1:e1-1__leaf1:e1-49"

	params := &network.NetemParams{
		DelayMs:     200,
		JitterMs:    20,
		LossPercent: 5.0,
	}
	server.faultState.Set(linkID, &network.LinkFaultState{
		LinkID: linkID,
		State:  "degraded",
		Netem:  params,
	})

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/test-fault-netem-get/links/"+linkID, nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var link linkResponse
	err := json.NewDecoder(w.Body).Decode(&link)
	require.NoError(t, err)
	assert.Equal(t, "degraded", link.State)
	require.NotNil(t, link.Netem)
	assert.Equal(t, 200, link.Netem.DelayMs)
	assert.Equal(t, 5.0, link.Netem.LossPercent)
}

func TestLinkFault_ActionClearNetem_ClearsState(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-clear", "topology-v073.json")
	linkID := "spine1:e1-1__leaf1:e1-49"

	params := &network.NetemParams{DelayMs: 100}
	server.faultState.Set(linkID, &network.LinkFaultState{
		LinkID: linkID,
		State:  "degraded",
		Netem:  params,
	})

	body := faultRequest{Action: "clear_netem"}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/test-fault-clear/links/"+linkID+"/fault",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	state := server.faultState.GetLinkState(linkID)
	assert.Equal(t, "up", state)
}

func TestLinkFault_InvalidAction(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-invalid", "topology-v073.json")

	body := faultRequest{Action: "explode"}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/test-fault-invalid/links/spine1:e1-1__leaf1:e1-49/fault",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLinkFault_NetemWithoutParams(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-netem-noparam", "topology-v073.json")

	body := faultRequest{Action: "netem"}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/test-fault-netem-noparam/links/spine1:e1-1__leaf1:e1-49/fault",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLinkFault_LinkNotFound(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-fault-linknotfound", "topology-v073.json")

	body := faultRequest{Action: "down"}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/projects/test-fault-linknotfound/links/nonexistent/fault",
		bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetLinks_AllLinksReflectFaultState(t *testing.T) {
	server, _ := setupLinksTestServer(t, "test-links-all-state", "topology-v073.json")
	linkID := "spine1:e1-1__leaf1:e1-49"

	// Pre-inject state
	server.faultState.Set(linkID, &network.LinkFaultState{
		LinkID: linkID,
		State:  "down",
	})

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/projects/test-links-all-state/links", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var links []linkResponse
	err := json.NewDecoder(w.Body).Decode(&links)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "down", links[0].State)
}
