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
	Racks            map[string]RackConfig            `yaml:"racks"`
	KindDefaults     map[string]KindConfig            `yaml:"kind_defaults"`
	Nodes            map[string]NodeConfig            `yaml:"nodes"`
	AutoMgmt         *AutoMgmtConfig                  `yaml:"auto_mgmt"`
	ExternalNodes    map[string]ExternalNodeConfig     `yaml:"external_nodes"`
	ExternalNetworks map[string]ExternalNetworkConfig  `yaml:"external_networks"`
	ExternalLinks    []ExternalLinkConfig              `yaml:"external_links"`
}

// AutoMgmtConfig controls auto-generation of management network display.
type AutoMgmtConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Position  string `yaml:"position"`  // "top" or "bottom" (default "bottom")
	Collapsed bool   `yaml:"collapsed"` // default true
}

// ExternalNodeConfig defines a non-clab device for visualization.
type ExternalNodeConfig struct {
	Label       string                    `yaml:"label"`
	Description string                    `yaml:"description"`
	Icon        string                    `yaml:"icon"`       // default "service"
	Interfaces  []string                  `yaml:"interfaces"`
	Placement   ExternalNodePlacement     `yaml:"placement"`
}

// ExternalNodePlacement defines where an external node is placed.
type ExternalNodePlacement struct {
	DC       string `yaml:"dc"`
	Rack     string `yaml:"rack"`
	RackUnit int    `yaml:"rack_unit"`
	Size     int    `yaml:"size"` // default 1
}

// ExternalNetworkConfig defines an external network (Internet, WAN, OOB, etc.).
type ExternalNetworkConfig struct {
	Label    string `yaml:"label"`
	Position string `yaml:"position"` // "top" or "bottom"
	DC       string `yaml:"dc"`
}

// ExternalLinkConfig defines a connection between any combination of clab nodes,
// external nodes, and external networks.
type ExternalLinkConfig struct {
	A ExternalLinkEndpointConfig `yaml:"a"`
	Z ExternalLinkEndpointConfig `yaml:"z"`
}

// ExternalLinkEndpointConfig identifies one side of an external link.
// Exactly one of Node, External, or Network must be set.
type ExternalLinkEndpointConfig struct {
	Node      string `yaml:"node"`
	External  string `yaml:"external"`
	Network   string `yaml:"network"`
	Interface string `yaml:"interface"`
}

// KindConfig holds kind-level configuration overrides.
type KindConfig struct {
	SSH *SSHCredentials `yaml:"ssh"`
}

// RackConfig holds rack-level configuration.
type RackConfig struct {
	DC    string `yaml:"dc"`
	Units int    `yaml:"units"` // default 42
}

// NodeConfig holds node-level visualization configuration.
type NodeConfig struct {
	Rack string          `yaml:"rack"`
	Unit int             `yaml:"unit"`
	Size int             `yaml:"size"` // default 1
	Role string          `yaml:"role"`
	SSH  *SSHCredentials `yaml:"ssh"`
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

	// Apply defaults for auto_mgmt
	if cfg.AutoMgmt != nil {
		if cfg.AutoMgmt.Position == "" {
			cfg.AutoMgmt.Position = "bottom"
		}
	}

	// Apply defaults for external nodes
	for name, en := range cfg.ExternalNodes {
		if en.Icon == "" {
			en.Icon = "service"
			cfg.ExternalNodes[name] = en
		}
		if en.Placement.Size == 0 {
			en.Placement.Size = 1
			cfg.ExternalNodes[name] = en
		}
	}

	// Apply defaults for external networks
	for name, net := range cfg.ExternalNetworks {
		if net.Position == "" {
			net.Position = "bottom"
			cfg.ExternalNetworks[name] = net
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
// This should be called after ApplyConfig and ApplyExternalConfig.
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

	// Build known DCs and racks from groups
	knownDCs := map[string]bool{}
	for _, dc := range topo.Groups.DCs {
		knownDCs[dc] = true
	}
	knownRacks := map[string]bool{}
	for _, racks := range topo.Groups.Racks {
		for _, rack := range racks {
			knownRacks[rack] = true
		}
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

	// Validate external nodes
	for _, en := range topo.ExternalNodes {
		g := en.Graph

		// Check: DC exists in known DCs (only if specified)
		if g.DC != "" && !knownDCs[g.DC] {
			warnings = append(warnings,
				fmt.Sprintf("external node %q: DC %q not found in topology", en.Name, g.DC))
		}

		// Check: rack exists in known racks (only if specified)
		if g.Rack != "" && !knownRacks[g.Rack] {
			warnings = append(warnings,
				fmt.Sprintf("external node %q: rack %q not found in topology", en.Name, g.Rack))
		}

		// Rack-placed external nodes: check unit overlap
		if g.Rack != "" && g.RackUnit > 0 {
			maxU := getRackUnits(g.Rack)
			unitStart := g.RackUnit
			unitEnd := g.RackUnit + g.RackUnitSize - 1

			if unitEnd > maxU {
				warnings = append(warnings,
					fmt.Sprintf("external node %q: unit range U%d–U%d exceeds rack %q height (%dU)",
						en.Name, unitStart, unitEnd, g.Rack, maxU))
			}

			rackPlacements[g.Rack] = append(rackPlacements[g.Rack], placement{
				start: unitStart,
				end:   unitEnd,
				node:  fmt.Sprintf("external:%s", en.Name),
			})
		}
	}

	// Validate external link endpoints
	nodeNames := map[string]bool{}
	for _, n := range topo.Nodes {
		nodeNames[n.Name] = true
	}
	externalNodeNames := map[string]bool{}
	externalNodeInterfaces := map[string]map[string]bool{}
	for _, en := range topo.ExternalNodes {
		externalNodeNames[en.Name] = true
		ifSet := map[string]bool{}
		for _, iface := range en.Interfaces {
			ifSet[iface] = true
		}
		externalNodeInterfaces[en.Name] = ifSet
	}
	networkNames := map[string]bool{}
	for _, net := range topo.ExternalNetworks {
		networkNames[net.Name] = true
	}

	for _, el := range topo.ExternalLinks {
		for _, side := range []struct {
			label string
			ep    ExternalEndpoint
		}{{"A", el.A}, {"Z", el.Z}} {
			ep := side.ep
			if ep.Node != "" && !nodeNames[ep.Node] {
				warnings = append(warnings,
					fmt.Sprintf("external link %q: %s-side node %q not found", el.ID, side.label, ep.Node))
			}
			if ep.External != "" && !externalNodeNames[ep.External] {
				warnings = append(warnings,
					fmt.Sprintf("external link %q: %s-side external node %q not found", el.ID, side.label, ep.External))
			}
			if ep.Network != "" && !networkNames[ep.Network] {
				warnings = append(warnings,
					fmt.Sprintf("external link %q: %s-side network %q not found", el.ID, side.label, ep.Network))
			}
			// Check interface exists for external nodes
			if ep.External != "" && ep.Interface != "" {
				if ifSet, ok := externalNodeInterfaces[ep.External]; ok {
					if !ifSet[ep.Interface] {
						warnings = append(warnings,
							fmt.Sprintf("external link %q: %s-side interface %q not found on external node %q",
								el.ID, side.label, ep.Interface, ep.External))
					}
				}
			}
		}
	}

	// Check: overlapping unit ranges within each rack (includes both clab + external nodes)
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
