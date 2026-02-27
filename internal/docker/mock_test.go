package docker

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
)

// MockDockerClient implements DockerClient for testing.
type MockDockerClient struct {
	Containers    []container.Summary
	InspectResult container.InspectResponse
	ExecCreateID  string
	EventsCh      chan events.Message
	ErrorsCh      chan error
	StartErr      error
	StopErr       error
	RestartErr    error
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return m.Containers, nil
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return m.InspectResult, nil
}

func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	return container.ExecCreateResponse{ID: m.ExecCreateID}, nil
}

func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error) {
	return types.HijackedResponse{
		Conn:   nil,
		Reader: nil,
	}, nil
}

func (m *MockDockerClient) ContainerExecResize(ctx context.Context, execID string, options container.ResizeOptions) error {
	return nil
}

func (m *MockDockerClient) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error) {
	return io.NopCloser(strings.NewReader("")), container.PathStat{}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return m.StartErr
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return m.StopErr
}

func (m *MockDockerClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	return m.RestartErr
}

func (m *MockDockerClient) Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error) {
	if m.EventsCh == nil {
		m.EventsCh = make(chan events.Message)
	}
	if m.ErrorsCh == nil {
		m.ErrorsCh = make(chan error)
	}
	return m.EventsCh, m.ErrorsCh
}
