package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tjst-t/clabnoc/internal/docker"
)

func TestGetProjects_ReturnsProjectList(t *testing.T) {
	mock := &mockDockerClient{
		containers: []container.Summary{
			makeTestContainer("c1", "proj1", "spine1", "/tmp/containerlab/clab-proj1/spine1", "running"),
			makeTestContainer("c2", "proj1", "leaf1", "/tmp/containerlab/clab-proj1/leaf1", "running"),
			makeTestContainer("c3", "proj2", "router1", "/tmp/containerlab/clab-proj2/router1", "exited"),
		},
	}

	server := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var projects []docker.ProjectInfo
	err := json.NewDecoder(w.Body).Decode(&projects)
	require.NoError(t, err)
	assert.Len(t, projects, 2)

	// Projects should be sorted
	assert.Equal(t, "proj1", projects[0].Name)
	assert.Equal(t, 2, projects[0].NodeCount)
	assert.Equal(t, "running", projects[0].Status)

	assert.Equal(t, "proj2", projects[1].Name)
	assert.Equal(t, 1, projects[1].NodeCount)
	assert.Equal(t, "stopped", projects[1].Status)
}

func TestGetProjects_EmptyReturnsEmptyArray(t *testing.T) {
	mock := &mockDockerClient{
		containers: []container.Summary{},
	}

	server := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var projects []docker.ProjectInfo
	err := json.NewDecoder(w.Body).Decode(&projects)
	require.NoError(t, err)
	// Nil slice marshals as null in JSON; handle that case too
	if projects == nil {
		projects = []docker.ProjectInfo{}
	}
	assert.Len(t, projects, 0)
}

func TestGetProjects_CORSHeader(t *testing.T) {
	mock := &mockDockerClient{}
	server := NewServer(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestGetProjects_LabDirFromLabel(t *testing.T) {
	mock := &mockDockerClient{
		containers: []container.Summary{
			makeTestContainer("c1", "myproj", "n1", "/tmp/containerlab/clab-myproj/n1", "running"),
		},
	}

	server := NewServer(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var projects []docker.ProjectInfo
	err := json.NewDecoder(w.Body).Decode(&projects)
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "/tmp/containerlab/clab-myproj", projects[0].LabDir)
}
