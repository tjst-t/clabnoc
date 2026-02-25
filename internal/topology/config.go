package topology

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Config represents .clabnoc.yml configuration.
type Config struct {
	Racks map[string]RackConfig `yaml:"racks"`
	Nodes map[string]NodeConfig `yaml:"nodes"`
}

// RackConfig holds rack-level configuration.
type RackConfig struct {
	DC    string `yaml:"dc"`
	Units int    `yaml:"units"` // default 42
}

// NodeConfig holds node-level visualization configuration.
type NodeConfig struct {
	Rack string `yaml:"rack"`
	Unit int    `yaml:"unit"`
	Size int    `yaml:"size"` // default 1
	Role string `yaml:"role"`
}

// FindConfigFile searches for .clabnoc.yml relative to labDir.
// Search order:
//  1. <labDir>/../<projectName>.clabnoc.yml  (clab YAML directory)
//  2. <labDir>/clabnoc.yml                    (inside labdir, for Docker mounts)
//
// Returns empty string if not found.
func FindConfigFile(labDir, projectName string) string {
	candidates := []string{
		filepath.Join(labDir, "..", projectName+".clabnoc.yml"),
		filepath.Join(labDir, "clabnoc.yml"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			abs, _ := filepath.Abs(path)
			slog.Debug("found config file", "path", abs)
			return abs
		}
	}
	return ""
}

// LoadConfigFile reads and parses a .clabnoc.yml file.
func LoadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Apply defaults
	for name, rack := range cfg.Racks {
		if rack.Units == 0 {
			rack.Units = 42
			cfg.Racks[name] = rack
		}
	}
	for name, node := range cfg.Nodes {
		if node.Size == 0 {
			node.Size = 1
			cfg.Nodes[name] = node
		}
	}

	return &cfg, nil
}

// ApplyConfig merges config into a parsed Topology.
// .clabnoc.yml settings override graph-* labels.
func ApplyConfig(topo *Topology, cfg *Config) {
	if cfg == nil {
		return
	}

	// Build rack→DC lookup from config
	rackDC := map[string]string{}
	for rackName, rc := range cfg.Racks {
		rackDC[rackName] = rc.DC
	}

	// Apply node overrides
	for i := range topo.Nodes {
		node := &topo.Nodes[i]
		nc, ok := cfg.Nodes[node.Name]
		if !ok {
			continue
		}

		if nc.Rack != "" {
			node.Graph.Rack = nc.Rack
			if dc, found := rackDC[nc.Rack]; found {
				node.Graph.DC = dc
			}
		}
		if nc.Unit != 0 {
			node.Graph.RackUnit = nc.Unit
		}
		if nc.Size != 0 {
			node.Graph.RackUnitSize = nc.Size
		}
		if nc.Role != "" {
			node.Graph.Role = nc.Role
			node.Graph.Icon = resolveIcon(node.Kind, node.Labels, isBMC(node.Image, node.Labels))
			// If role is explicitly set via config, re-resolve icon with updated role label
			if nc.Role != "" {
				labelsWithRole := make(map[string]string, len(node.Labels)+1)
				for k, v := range node.Labels {
					labelsWithRole[k] = v
				}
				labelsWithRole["graph-role"] = nc.Role
				node.Graph.Icon = resolveIcon(node.Kind, labelsWithRole, isBMC(node.Image, node.Labels))
			}
		}
	}

	// Apply rack units to Groups
	if topo.Groups.RackUnits == nil {
		topo.Groups.RackUnits = make(map[string]int)
	}
	for rackName, rc := range cfg.Racks {
		topo.Groups.RackUnits[rackName] = rc.Units
	}

	// Rebuild Groups.DCs and Groups.Racks from merged node data
	rebuildGroups(topo)
}

// ValidateLayout checks for layout inconsistencies and returns warnings.
// This should be called after ApplyConfig (or after Parse if no config).
func ValidateLayout(topo *Topology) []string {
	var warnings []string

	// Build rack units map (rack name → U count, default 42)
	rackUnits := map[string]int{}
	for rack, units := range topo.Groups.RackUnits {
		rackUnits[rack] = units
	}
	getRackUnits := func(rack string) int {
		if u, ok := rackUnits[rack]; ok {
			return u
		}
		return 42
	}

	// Track occupied U ranges per rack: rack → list of (start, end, nodeName)
	type placement struct {
		start, end int
		node       string
	}
	rackPlacements := map[string][]placement{}

	for _, node := range topo.Nodes {
		g := node.Graph

		// Check: node missing rack placement
		if g.DC == "" || g.Rack == "" || g.RackUnit == 0 {
			warnings = append(warnings,
				fmt.Sprintf("node %q: missing rack placement (dc=%q, rack=%q, unit=%d)",
					node.Name, g.DC, g.Rack, g.RackUnit))
			continue
		}

		rack := g.Rack
		maxU := getRackUnits(rack)
		unitStart := g.RackUnit
		unitEnd := g.RackUnit + g.RackUnitSize - 1

		// Check: unit exceeds rack height
		if unitEnd > maxU {
			warnings = append(warnings,
				fmt.Sprintf("node %q: unit range U%d–U%d exceeds rack %q height (%dU)",
					node.Name, unitStart, unitEnd, rack, maxU))
		}

		// Collect placement for overlap check
		rackPlacements[rack] = append(rackPlacements[rack], placement{
			start: unitStart,
			end:   unitEnd,
			node:  node.Name,
		})
	}

	// Check: overlapping unit ranges within each rack
	for rack, placements := range rackPlacements {
		for i := 0; i < len(placements); i++ {
			for j := i + 1; j < len(placements); j++ {
				a := placements[i]
				b := placements[j]
				// Overlap: ranges [a.start, a.end] and [b.start, b.end] intersect
				if a.start <= b.end && b.start <= a.end {
					warnings = append(warnings,
						fmt.Sprintf("rack %q: nodes %q (U%d–U%d) and %q (U%d–U%d) overlap",
							rack, a.node, a.start, a.end, b.node, b.start, b.end))
				}
			}
		}
	}

	return warnings
}

// rebuildGroups reconstructs DCs and Racks from current node data.
func rebuildGroups(topo *Topology) {
	dcSet := map[string]bool{}
	racksByDC := map[string]map[string]bool{}

	for _, node := range topo.Nodes {
		dc := node.Graph.DC
		rack := node.Graph.Rack
		if dc == "" {
			continue
		}
		if !dcSet[dc] {
			dcSet[dc] = true
			racksByDC[dc] = map[string]bool{}
		}
		if rack != "" {
			racksByDC[dc][rack] = true
		}
	}

	topo.Groups.DCs = make([]string, 0, len(dcSet))
	for dc := range dcSet {
		topo.Groups.DCs = append(topo.Groups.DCs, dc)
	}
	sort.Strings(topo.Groups.DCs)

	topo.Groups.Racks = make(map[string][]string)
	for dc, racksMap := range racksByDC {
		racks := make([]string, 0, len(racksMap))
		for rack := range racksMap {
			racks = append(racks, rack)
		}
		sort.Strings(racks)
		topo.Groups.Racks[dc] = racks
	}
}
