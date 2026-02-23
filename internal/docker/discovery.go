package docker

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// ProjectInfo holds metadata about a discovered clab project
type ProjectInfo struct {
	Name      string `json:"name"`
	NodeCount int    `json:"nodes"`
	Status    string `json:"status"` // "running", "partial", "stopped"
	LabDir    string `json:"labdir"`
}

// Discoverer discovers clab projects from Docker
type Discoverer struct {
	client DockerClient
}

// NewDiscoverer creates a new Discoverer
func NewDiscoverer(client DockerClient) *Discoverer {
	return &Discoverer{client: client}
}

// ListProjects returns all clab projects found in Docker
// Filter by label "containerlab", group by project name
// Status: "running"=all running, "stopped"=none running, "partial"=some running
// LabDir: derived from clab-node-lab-dir label or default to /tmp/containerlab/clab-{name}
func (d *Discoverer) ListProjects(ctx context.Context) ([]ProjectInfo, error) {
	f := filters.NewArgs(filters.Arg("label", "containerlab"))
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	type projectData struct {
		totalCount   int
		runningCount int
		labDir       string
	}

	projects := map[string]*projectData{}
	for _, c := range containers {
		projName := c.Labels["containerlab"]
		if projName == "" {
			continue
		}

		if _, ok := projects[projName]; !ok {
			projects[projName] = &projectData{}
		}
		pd := projects[projName]
		pd.totalCount++

		if c.State == "running" {
			pd.runningCount++
		}

		// Derive lab dir from node lab dir label
		if pd.labDir == "" {
			nodeLabDir := c.Labels["clab-node-lab-dir"]
			if nodeLabDir != "" {
				pd.labDir = filepath.Dir(nodeLabDir)
			}
		}
	}

	result := make([]ProjectInfo, 0, len(projects))
	for name, pd := range projects {
		status := "stopped"
		if pd.runningCount == pd.totalCount && pd.totalCount > 0 {
			status = "running"
		} else if pd.runningCount > 0 {
			status = "partial"
		}

		labDir := pd.labDir
		if labDir == "" {
			labDir = "/tmp/containerlab/clab-" + name
		}

		result = append(result, ProjectInfo{
			Name:      name,
			NodeCount: pd.totalCount,
			Status:    status,
			LabDir:    labDir,
		})
	}

	// Sort for deterministic output
	sortProjects(result)
	return result, nil
}

// GetContainerByNode finds container by clab node name within a project
// Matches by label: containerlab={project} AND clab-node-name={node}
func (d *Discoverer) GetContainerByNode(ctx context.Context, project, node string) (container.Summary, error) {
	f := filters.NewArgs(
		filters.Arg("label", "containerlab="+project),
		filters.Arg("label", "clab-node-name="+node),
	)
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return container.Summary{}, fmt.Errorf("list containers for node %s/%s: %w", project, node, err)
	}
	if len(containers) == 0 {
		return container.Summary{}, fmt.Errorf("container not found for node %s in project %s", node, project)
	}
	return containers[0], nil
}

// GetProjectLabDir returns the labdir for a project
func (d *Discoverer) GetProjectLabDir(ctx context.Context, name string) (string, error) {
	f := filters.NewArgs(filters.Arg("label", "containerlab="+name))
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return "", fmt.Errorf("list containers for project %s: %w", name, err)
	}

	for _, c := range containers {
		nodeLabDir := c.Labels["clab-node-lab-dir"]
		if nodeLabDir != "" {
			return filepath.Dir(nodeLabDir), nil
		}
	}

	// Default path
	return "/tmp/containerlab/clab-" + name, nil
}

// sortProjects sorts projects by name for deterministic output
func sortProjects(projects []ProjectInfo) {
	for i := 1; i < len(projects); i++ {
		for j := i; j > 0 && projects[j].Name < projects[j-1].Name; j-- {
			projects[j], projects[j-1] = projects[j-1], projects[j]
		}
	}
}

// GetContainersByProject returns all containers for a project
func (d *Discoverer) GetContainersByProject(ctx context.Context, project string) ([]container.Summary, error) {
	f := filters.NewArgs(filters.Arg("label", "containerlab="+project))
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers for project %s: %w", project, err)
	}
	return containers, nil
}

// containerStateName converts Docker container state to a readable status
func containerStateName(state string) string {
	return strings.ToLower(state)
}
