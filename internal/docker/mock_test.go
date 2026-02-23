package docker

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

// MockDockerClient is a mock implementation of DockerClient for testing
type MockDockerClient struct {
	containers []container.Summary
	inspectFn  func(id string) (container.InspectResponse, error)
	listErr    error
}

func (m *MockDockerClient) ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	// Filter containers based on label filters
	labelFilters := opts.Filters.Get("label")
	if len(labelFilters) == 0 {
		return m.containers, nil
	}

	var result []container.Summary
	for _, c := range m.containers {
		if containerMatchesLabelFilters(c, labelFilters) {
			result = append(result, c)
		}
	}
	return result, nil
}

// containerMatchesLabelFilters returns true if the container has all the required labels
func containerMatchesLabelFilters(c container.Summary, labelFilters []string) bool {
	for _, filter := range labelFilters {
		if strings.Contains(filter, "=") {
			// key=value match
			parts := strings.SplitN(filter, "=", 2)
			key, value := parts[0], parts[1]
			if c.Labels[key] != value {
				return false
			}
		} else {
			// key exists match
			if _, ok := c.Labels[filter]; !ok {
				return false
			}
		}
	}
	return true
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, id string) (container.InspectResponse, error) {
	if m.inspectFn != nil {
		return m.inspectFn(id)
	}
	return container.InspectResponse{}, nil
}

func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, id string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	return container.ExecCreateResponse{ID: "test-exec-id"}, nil
}

func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, id string, config container.ExecAttachOptions) (types.HijackedResponse, error) {
	return types.HijackedResponse{}, nil
}

func (m *MockDockerClient) CopyFromContainer(ctx context.Context, id, path string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(nil), container.PathStat{}, nil
}

func (m *MockDockerClient) Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
	ch := make(chan events.Message)
	errCh := make(chan error)
	return ch, errCh
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, id string, opts container.StartOptions) error {
	return nil
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, id string, opts container.StopOptions) error {
	return nil
}

// makeContainer creates a test container.Summary
func makeContainer(id, name, project, nodeName, nodeLabDir, state string) container.Summary {
	labels := map[string]string{
		"containerlab":      project,
		"clab-node-name":    nodeName,
		"clab-node-lab-dir": nodeLabDir,
	}
	return container.Summary{
		ID:     id,
		Names:  []string{"/" + name},
		State:  state,
		Status: state,
		Labels: labels,
	}
}

// makeContainerWithFilters creates a test container for use with filters
func makeContainerWithFilters(id, name string, labels map[string]string, state string) container.Summary {
	return container.Summary{
		ID:     id,
		Names:  []string{"/" + name},
		State:  state,
		Status: state,
		Labels: labels,
	}
}

// filtersFromArgs is a helper to create filters.Args from key-value pairs
func filtersFromArgs(args ...filters.KeyValuePair) filters.Args {
	return filters.NewArgs(args...)
}
