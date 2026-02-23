package api

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
)

// mockDockerClient implements docker.DockerClient for testing
type mockDockerClient struct {
	containers []container.Summary
	listErr    error
	eventsFn   func(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error)
}

func (m *mockDockerClient) ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	// Filter by label if Filters are set
	labelFilters := opts.Filters.Get("label")
	if len(labelFilters) == 0 {
		return m.containers, nil
	}

	var result []container.Summary
	for _, c := range m.containers {
		if matchesAllLabels(c, labelFilters) {
			result = append(result, c)
		}
	}
	return result, nil
}

func matchesAllLabels(c container.Summary, labelFilters []string) bool {
	for _, f := range labelFilters {
		if idx := indexByte(f, '='); idx >= 0 {
			key := f[:idx]
			val := f[idx+1:]
			if c.Labels[key] != val {
				return false
			}
		} else {
			if _, ok := c.Labels[f]; !ok {
				return false
			}
		}
	}
	return true
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func (m *mockDockerClient) ContainerInspect(ctx context.Context, id string) (container.InspectResponse, error) {
	return container.InspectResponse{}, nil
}

func (m *mockDockerClient) ContainerExecCreate(ctx context.Context, id string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	return container.ExecCreateResponse{ID: "exec-id"}, nil
}

func (m *mockDockerClient) ContainerExecAttach(ctx context.Context, id string, config container.ExecAttachOptions) (types.HijackedResponse, error) {
	return types.HijackedResponse{}, nil
}

func (m *mockDockerClient) CopyFromContainer(ctx context.Context, id, path string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(nil), container.PathStat{}, nil
}

func (m *mockDockerClient) Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
	if m.eventsFn != nil {
		return m.eventsFn(ctx, opts)
	}
	ch := make(chan events.Message)
	errCh := make(chan error)
	return ch, errCh
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, id string, opts container.StartOptions) error {
	return nil
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, id string, opts container.StopOptions) error {
	return nil
}

// makeTestContainer creates a test container.Summary for API tests
func makeTestContainer(id, project, nodeName, nodeLabDir, state string) container.Summary {
	return container.Summary{
		ID:    id,
		Names: []string{"/clab-" + project + "-" + nodeName},
		State: state,
		Labels: map[string]string{
			"containerlab":      project,
			"clab-node-name":    nodeName,
			"clab-node-lab-dir": nodeLabDir,
		},
	}
}
