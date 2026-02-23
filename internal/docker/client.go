package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// DockerClient abstracts the Docker SDK operations needed by clabnoc.
type DockerClient interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error)
	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error
	Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error)
}

// RealClient wraps the real Docker SDK client.
type RealClient struct {
	cli *client.Client
}

// NewRealClient creates a new Docker client using the environment configuration.
func NewRealClient() (*RealClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &RealClient{cli: cli}, nil
}

func (r *RealClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return r.cli.ContainerList(ctx, options)
}

func (r *RealClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return r.cli.ContainerInspect(ctx, containerID)
}

func (r *RealClient) ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	return r.cli.ContainerExecCreate(ctx, containerID, config)
}

func (r *RealClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error) {
	return r.cli.ContainerExecAttach(ctx, execID, config)
}

func (r *RealClient) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error) {
	return r.cli.CopyFromContainer(ctx, containerID, srcPath)
}

func (r *RealClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return r.cli.ContainerStart(ctx, containerID, options)
}

func (r *RealClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return r.cli.ContainerStop(ctx, containerID, options)
}

func (r *RealClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	return r.cli.ContainerRestart(ctx, containerID, options)
}

func (r *RealClient) Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error) {
	return r.cli.Events(ctx, options)
}

// ClabContainerFilter returns a ListOptions filter for containerlab containers.
func ClabContainerFilter() container.ListOptions {
	f := filters.NewArgs()
	f.Add("label", "containerlab")
	return container.ListOptions{All: true, Filters: f}
}
