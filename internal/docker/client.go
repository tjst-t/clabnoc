package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	dockerclient "github.com/docker/docker/client"
)

// DockerClient is an interface for Docker operations (enables mocking in tests)
type DockerClient interface {
	ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error)
	ContainerInspect(ctx context.Context, id string) (container.InspectResponse, error)
	ContainerExecCreate(ctx context.Context, id string, config container.ExecOptions) (container.ExecCreateResponse, error)
	ContainerExecAttach(ctx context.Context, id string, config container.ExecAttachOptions) (types.HijackedResponse, error)
	CopyFromContainer(ctx context.Context, id, path string) (io.ReadCloser, container.PathStat, error)
	Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error)
	ContainerStart(ctx context.Context, id string, opts container.StartOptions) error
	ContainerStop(ctx context.Context, id string, opts container.StopOptions) error
}

// RealClient wraps the Docker SDK client
type RealClient struct {
	cli *dockerclient.Client
}

// NewClient creates a new Docker client from environment
func NewClient() (*RealClient, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &RealClient{cli: cli}, nil
}

func (c *RealClient) ContainerList(ctx context.Context, opts container.ListOptions) ([]container.Summary, error) {
	return c.cli.ContainerList(ctx, opts)
}

func (c *RealClient) ContainerInspect(ctx context.Context, id string) (container.InspectResponse, error) {
	return c.cli.ContainerInspect(ctx, id)
}

func (c *RealClient) ContainerExecCreate(ctx context.Context, id string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	return c.cli.ContainerExecCreate(ctx, id, config)
}

func (c *RealClient) ContainerExecAttach(ctx context.Context, id string, config container.ExecAttachOptions) (types.HijackedResponse, error) {
	return c.cli.ContainerExecAttach(ctx, id, config)
}

func (c *RealClient) CopyFromContainer(ctx context.Context, id, path string) (io.ReadCloser, container.PathStat, error) {
	return c.cli.CopyFromContainer(ctx, id, path)
}

func (c *RealClient) Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error) {
	return c.cli.Events(ctx, opts)
}

func (c *RealClient) ContainerStart(ctx context.Context, id string, opts container.StartOptions) error {
	return c.cli.ContainerStart(ctx, id, opts)
}

func (c *RealClient) ContainerStop(ctx context.Context, id string, opts container.StopOptions) error {
	return c.cli.ContainerStop(ctx, id, opts)
}
