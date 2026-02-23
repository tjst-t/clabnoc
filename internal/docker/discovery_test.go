package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProjects_GroupsByProjectName(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-proj1-spine1", "proj1", "spine1", "/tmp/containerlab/clab-proj1/spine1", "running"),
			makeContainer("c2", "clab-proj1-leaf1", "proj1", "leaf1", "/tmp/containerlab/clab-proj1/leaf1", "running"),
			makeContainer("c3", "clab-proj2-router1", "proj2", "router1", "/tmp/containerlab/clab-proj2/router1", "running"),
		},
	}

	d := NewDiscoverer(mock)
	projects, err := d.ListProjects(context.Background())
	require.NoError(t, err)
	assert.Len(t, projects, 2)

	// Projects should be sorted by name
	assert.Equal(t, "proj1", projects[0].Name)
	assert.Equal(t, 2, projects[0].NodeCount)
	assert.Equal(t, "proj2", projects[1].Name)
	assert.Equal(t, 1, projects[1].NodeCount)
}

func TestListProjects_StatusRunning(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-proj1-n1", "proj1", "n1", "/tmp/containerlab/clab-proj1/n1", "running"),
			makeContainer("c2", "clab-proj1-n2", "proj1", "n2", "/tmp/containerlab/clab-proj1/n2", "running"),
		},
	}

	d := NewDiscoverer(mock)
	projects, err := d.ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "running", projects[0].Status)
}

func TestListProjects_StatusStopped(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-proj1-n1", "proj1", "n1", "/tmp/containerlab/clab-proj1/n1", "exited"),
			makeContainer("c2", "clab-proj1-n2", "proj1", "n2", "/tmp/containerlab/clab-proj1/n2", "exited"),
		},
	}

	d := NewDiscoverer(mock)
	projects, err := d.ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "stopped", projects[0].Status)
}

func TestListProjects_StatusPartial(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-proj1-n1", "proj1", "n1", "/tmp/containerlab/clab-proj1/n1", "running"),
			makeContainer("c2", "clab-proj1-n2", "proj1", "n2", "/tmp/containerlab/clab-proj1/n2", "exited"),
		},
	}

	d := NewDiscoverer(mock)
	projects, err := d.ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "partial", projects[0].Status)
}

func TestListProjects_LabDirFromLabel(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-myproj-n1", "myproj", "n1", "/tmp/containerlab/clab-myproj/n1", "running"),
		},
	}

	d := NewDiscoverer(mock)
	projects, err := d.ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "/tmp/containerlab/clab-myproj", projects[0].LabDir)
}

func TestListProjects_LabDirDefault(t *testing.T) {
	// Container without clab-node-lab-dir label
	c := makeContainer("c1", "clab-myproj-n1", "myproj", "n1", "", "running")
	delete(c.Labels, "clab-node-lab-dir")

	mock := &MockDockerClient{
		containers: []container.Summary{c},
	}

	d := NewDiscoverer(mock)
	projects, err := d.ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "/tmp/containerlab/clab-myproj", projects[0].LabDir)
}

func TestListProjects_EmptyReturnsEmpty(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{},
	}

	d := NewDiscoverer(mock)
	projects, err := d.ListProjects(context.Background())
	require.NoError(t, err)
	assert.Len(t, projects, 0)
}

func TestListProjects_DockerError(t *testing.T) {
	mock := &MockDockerClient{
		listErr: assert.AnError,
	}

	d := NewDiscoverer(mock)
	_, err := d.ListProjects(context.Background())
	assert.Error(t, err)
}

func TestGetContainerByNode_Found(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-proj1-spine1", "proj1", "spine1", "/tmp/containerlab/clab-proj1/spine1", "running"),
			makeContainer("c2", "clab-proj1-leaf1", "proj1", "leaf1", "/tmp/containerlab/clab-proj1/leaf1", "running"),
		},
	}

	d := NewDiscoverer(mock)
	c, err := d.GetContainerByNode(context.Background(), "proj1", "spine1")
	require.NoError(t, err)
	assert.Equal(t, "c1", c.ID)
}

func TestGetContainerByNode_NotFound(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-proj1-spine1", "proj1", "spine1", "/tmp/containerlab/clab-proj1/spine1", "running"),
		},
	}

	d := NewDiscoverer(mock)
	_, err := d.GetContainerByNode(context.Background(), "proj1", "nonexistent")
	assert.Error(t, err)
}

func TestGetProjectLabDir_FromLabel(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{
			makeContainer("c1", "clab-myproj-n1", "myproj", "n1", "/tmp/containerlab/clab-myproj/n1", "running"),
		},
	}

	d := NewDiscoverer(mock)
	labDir, err := d.GetProjectLabDir(context.Background(), "myproj")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/containerlab/clab-myproj", labDir)
}

func TestGetProjectLabDir_DefaultPath(t *testing.T) {
	// Container without clab-node-lab-dir label
	c := makeContainer("c1", "clab-myproj-n1", "myproj", "n1", "", "running")
	delete(c.Labels, "clab-node-lab-dir")

	mock := &MockDockerClient{
		containers: []container.Summary{c},
	}

	d := NewDiscoverer(mock)
	labDir, err := d.GetProjectLabDir(context.Background(), "myproj")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/containerlab/clab-myproj", labDir)
}

func TestGetProjectLabDir_NoContainers(t *testing.T) {
	mock := &MockDockerClient{
		containers: []container.Summary{},
	}

	d := NewDiscoverer(mock)
	labDir, err := d.GetProjectLabDir(context.Background(), "myproj")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/containerlab/clab-myproj", labDir)
}
