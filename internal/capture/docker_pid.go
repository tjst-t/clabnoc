package capture

import (
	"context"
	"fmt"

	"github.com/tjst-t/clabnoc/internal/docker"
)

// DockerPIDProvider retrieves container PIDs using the Docker API.
type DockerPIDProvider struct {
	client docker.DockerClient
}

// NewDockerPIDProvider creates a new DockerPIDProvider.
func NewDockerPIDProvider(client docker.DockerClient) *DockerPIDProvider {
	return &DockerPIDProvider{client: client}
}

// GetPID returns the PID of the container's init process.
func (p *DockerPIDProvider) GetPID(ctx context.Context, containerID string) (int, error) {
	info, err := p.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, fmt.Errorf("inspecting container %s: %w", containerID, err)
	}
	if info.State == nil || info.State.Pid == 0 {
		return 0, fmt.Errorf("container %s is not running", containerID)
	}
	return info.State.Pid, nil
}
