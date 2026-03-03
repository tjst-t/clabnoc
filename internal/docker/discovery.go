package docker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/tjst-t/clabnoc/internal/topology"
)

// ProjectInfo holds information about a discovered clab project.
type ProjectInfo struct {
	Name       string                   `json:"name"`
	NodeCount  int                      `json:"nodes"`
	Status     string                   `json:"status"`
	LabDir     string                   `json:"labdir"`
	Containers []container.Summary      `json:"-"`
}

// DiscoverProjects finds all containerlab projects via Docker API.
func DiscoverProjects(ctx context.Context, cli DockerClient) ([]ProjectInfo, error) {
	containers, err := cli.ContainerList(ctx, ClabContainerFilter())
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	projectMap := map[string]*ProjectInfo{}
	for _, c := range containers {
		proj := c.Labels["containerlab"]
		if proj == "" {
			continue
		}

		if _, ok := projectMap[proj]; !ok {
			labDir := inferLabDir(proj, c.Labels)
			projectMap[proj] = &ProjectInfo{
				Name:   proj,
				LabDir: labDir,
			}
		}

		p := projectMap[proj]
		p.NodeCount++
		p.Containers = append(p.Containers, c)
	}

	projects := make([]ProjectInfo, 0, len(projectMap))
	for _, p := range projectMap {
		p.Status = computeStatus(p.Containers)
		projects = append(projects, *p)
	}

	return projects, nil
}

// TopologyWithConfig holds a parsed topology and its associated config.
type TopologyWithConfig struct {
	Topology *topology.Topology
	Config   *topology.Config
}

// GetProjectTopologyWithConfig loads the topology and config for a specific project.
func GetProjectTopologyWithConfig(ctx context.Context, cli DockerClient, projectName string) (*TopologyWithConfig, error) {
	projects, err := DiscoverProjects(ctx, cli)
	if err != nil {
		return nil, err
	}

	var project *ProjectInfo
	for _, p := range projects {
		if p.Name == projectName {
			project = &p
			break
		}
	}
	if project == nil {
		return nil, fmt.Errorf("project %q not found", projectName)
	}

	topoData, err := loadTopologyData(project.LabDir)
	if err != nil {
		return nil, fmt.Errorf("loading topology data: %w", err)
	}

	raw, err := topology.ParseRaw(topoData)
	if err != nil {
		return nil, fmt.Errorf("parsing topology: %w", err)
	}

	topo := topology.Convert(raw)

	var cfg *topology.Config

	// Load .clabnoc.yml config if available
	if cfgPath := topology.FindConfigFile(project.LabDir, projectName); cfgPath != "" {
		loadedCfg, cfgErr := topology.LoadConfigFile(cfgPath)
		if cfgErr != nil {
			slog.Warn("failed to load .clabnoc.yml", "path", cfgPath, "error", cfgErr)
		} else {
			slog.Info("applying .clabnoc.yml config", "path", cfgPath)
			cfg = loadedCfg
			topology.ApplyConfig(topo, cfg)
			topology.ApplyExternalConfig(topo, cfg, raw)
		}
	}

	enrichWithContainerInfo(topo, project.Containers)

	// Validate layout and attach warnings
	if warns := topology.ValidateLayout(topo); len(warns) > 0 {
		topo.Warnings = warns
		for _, w := range warns {
			slog.Warn("layout validation", "project", projectName, "warning", w)
		}
	}

	return &TopologyWithConfig{Topology: topo, Config: cfg}, nil
}

// GetProjectTopology loads the topology for a specific project.
func GetProjectTopology(ctx context.Context, cli DockerClient, projectName string) (*topology.Topology, error) {
	result, err := GetProjectTopologyWithConfig(ctx, cli, projectName)
	if err != nil {
		return nil, err
	}
	return result.Topology, nil
}

// FindContainerByNode finds a container by project and node name.
func FindContainerByNode(ctx context.Context, cli DockerClient, projectName, nodeName string) (*container.Summary, error) {
	containers, err := cli.ContainerList(ctx, ClabContainerFilter())
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	for _, c := range containers {
		if c.Labels["containerlab"] == projectName && c.Labels["clab-node-name"] == nodeName {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("container for node %q in project %q not found", nodeName, projectName)
}

func inferLabDir(projectName string, labels map[string]string) string {
	if labDir := labels["clab-node-lab-dir"]; labDir != "" {
		return filepath.Dir(labDir)
	}
	return fmt.Sprintf("/tmp/containerlab/clab-%s", projectName)
}

func computeStatus(containers []container.Summary) string {
	running := 0
	for _, c := range containers {
		if c.State == "running" {
			running++
		}
	}
	switch {
	case running == len(containers):
		return "running"
	case running == 0:
		return "stopped"
	default:
		return "partial"
	}
}

func loadTopologyData(labDir string) ([]byte, error) {
	topoPath := filepath.Join(labDir, "topology-data.json")
	data, err := os.ReadFile(topoPath)
	if err != nil {
		slog.Warn("failed to read topology-data.json from filesystem", "path", topoPath, "error", err)
		return nil, fmt.Errorf("reading %s: %w", topoPath, err)
	}
	return data, nil
}

func enrichWithContainerInfo(topo *topology.Topology, containers []container.Summary) {
	containerByNode := map[string]container.Summary{}
	for _, c := range containers {
		if name := c.Labels["clab-node-name"]; name != "" {
			containerByNode[name] = c
		}
	}

	for i := range topo.Nodes {
		node := &topo.Nodes[i]
		if c, ok := containerByNode[node.Name]; ok {
			node.ContainerID = c.ID
			if c.State == "running" {
				node.Status = "running"
			} else {
				node.Status = "stopped"
			}
		}
	}
}
