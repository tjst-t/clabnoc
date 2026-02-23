package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestDiscoverProjects(t *testing.T) {
	mock := &MockDockerClient{
		Containers: []container.Summary{
			{
				ID:    "abc123",
				State: "running",
				Labels: map[string]string{
					"containerlab":  "project-a",
					"clab-node-name": "spine1",
				},
			},
			{
				ID:    "def456",
				State: "running",
				Labels: map[string]string{
					"containerlab":  "project-a",
					"clab-node-name": "leaf1",
				},
			},
			{
				ID:    "ghi789",
				State: "running",
				Labels: map[string]string{
					"containerlab":  "project-b",
					"clab-node-name": "router1",
				},
			},
		},
	}

	projects, err := DiscoverProjects(context.Background(), mock)
	if err != nil {
		t.Fatalf("DiscoverProjects failed: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	projectMap := map[string]ProjectInfo{}
	for _, p := range projects {
		projectMap[p.Name] = p
	}

	pa := projectMap["project-a"]
	if pa.NodeCount != 2 {
		t.Errorf("project-a: expected 2 nodes, got %d", pa.NodeCount)
	}
	if pa.Status != "running" {
		t.Errorf("project-a: expected status running, got %s", pa.Status)
	}

	pb := projectMap["project-b"]
	if pb.NodeCount != 1 {
		t.Errorf("project-b: expected 1 node, got %d", pb.NodeCount)
	}
}

func TestDiscoverProjectsEmpty(t *testing.T) {
	mock := &MockDockerClient{Containers: []container.Summary{}}
	projects, err := DiscoverProjects(context.Background(), mock)
	if err != nil {
		t.Fatalf("DiscoverProjects failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestComputeStatus(t *testing.T) {
	tests := []struct {
		name       string
		containers []container.Summary
		want       string
	}{
		{
			"all running",
			[]container.Summary{
				{State: "running"},
				{State: "running"},
			},
			"running",
		},
		{
			"all stopped",
			[]container.Summary{
				{State: "exited"},
				{State: "exited"},
			},
			"stopped",
		},
		{
			"partial",
			[]container.Summary{
				{State: "running"},
				{State: "exited"},
			},
			"partial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeStatus(tt.containers)
			if got != tt.want {
				t.Errorf("computeStatus = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindContainerByNode(t *testing.T) {
	mock := &MockDockerClient{
		Containers: []container.Summary{
			{
				ID: "abc123",
				Labels: map[string]string{
					"containerlab":  "myproject",
					"clab-node-name": "spine1",
				},
			},
			{
				ID: "def456",
				Labels: map[string]string{
					"containerlab":  "myproject",
					"clab-node-name": "leaf1",
				},
			},
		},
	}

	c, err := FindContainerByNode(context.Background(), mock, "myproject", "spine1")
	if err != nil {
		t.Fatalf("FindContainerByNode failed: %v", err)
	}
	if c.ID != "abc123" {
		t.Errorf("expected container abc123, got %s", c.ID)
	}

	_, err = FindContainerByNode(context.Background(), mock, "myproject", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestInferLabDir(t *testing.T) {
	tests := []struct {
		name    string
		project string
		labels  map[string]string
		want    string
	}{
		{
			"from label",
			"test",
			map[string]string{"clab-node-lab-dir": "/tmp/containerlab/clab-test/spine1"},
			"/tmp/containerlab/clab-test",
		},
		{
			"default",
			"test",
			map[string]string{},
			"/tmp/containerlab/clab-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferLabDir(tt.project, tt.labels)
			if got != tt.want {
				t.Errorf("inferLabDir = %q, want %q", got, tt.want)
			}
		})
	}
}
