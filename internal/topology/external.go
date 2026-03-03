package topology

import (
	"fmt"
	"sort"
)

// ApplyExternalConfig resolves external nodes, networks, links, and auto_mgmt
// from the .clabnoc.yml config and adds them to the topology.
func ApplyExternalConfig(topo *Topology, cfg *Config, raw *RawTopology) {
	if cfg == nil {
		return
	}

	// Build rack→DC lookup from config
	rackDC := map[string]string{}
	for rackName, rc := range cfg.Racks {
		rackDC[rackName] = rc.DC
	}

	applyExternalNodes(topo, cfg, rackDC)
	applyExternalNetworks(topo, cfg)
	applyExternalLinks(topo, cfg)
	applyAutoMgmt(topo, cfg, raw)
}

// applyExternalNodes converts cfg.ExternalNodes → topo.ExternalNodes.
func applyExternalNodes(topo *Topology, cfg *Config, rackDC map[string]string) {
	if len(cfg.ExternalNodes) == 0 {
		return
	}

	// Sort names for deterministic output
	names := make([]string, 0, len(cfg.ExternalNodes))
	for name := range cfg.ExternalNodes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		enc := cfg.ExternalNodes[name]

		dc := enc.Placement.DC
		// If rack is specified, resolve DC from rack→DC lookup
		if dc == "" && enc.Placement.Rack != "" {
			if d, ok := rackDC[enc.Placement.Rack]; ok {
				dc = d
			}
		}

		graph := GraphInfo{
			DC:           dc,
			Rack:         enc.Placement.Rack,
			RackUnit:     enc.Placement.RackUnit,
			RackUnitSize: enc.Placement.Size,
			Icon:         enc.Icon,
		}

		topo.ExternalNodes = append(topo.ExternalNodes, ExternalNode{
			Name:        name,
			Label:       enc.Label,
			Description: enc.Description,
			Icon:        enc.Icon,
			Interfaces:  enc.Interfaces,
			Graph:       graph,
			External:    true,
		})
	}
}

// applyExternalNetworks converts cfg.ExternalNetworks → topo.ExternalNetworks.
func applyExternalNetworks(topo *Topology, cfg *Config) {
	if len(cfg.ExternalNetworks) == 0 {
		return
	}

	// Sort names for deterministic output
	names := make([]string, 0, len(cfg.ExternalNetworks))
	for name := range cfg.ExternalNetworks {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		enc := cfg.ExternalNetworks[name]
		topo.ExternalNetworks = append(topo.ExternalNetworks, ExternalNetwork{
			Name:     name,
			Label:    enc.Label,
			Position: enc.Position,
			DC:       enc.DC,
		})
	}
}

// applyExternalLinks converts cfg.ExternalLinks → topo.ExternalLinks.
func applyExternalLinks(topo *Topology, cfg *Config) {
	for _, elc := range cfg.ExternalLinks {
		aRef := endpointRef(elc.A)
		zRef := endpointRef(elc.Z)
		id := fmt.Sprintf("ext:%s:%s__%s:%s", aRef, elc.A.Interface, zRef, elc.Z.Interface)

		topo.ExternalLinks = append(topo.ExternalLinks, ExternalLink{
			ID: id,
			A: ExternalEndpoint{
				Node:      elc.A.Node,
				External:  elc.A.External,
				Network:   elc.A.Network,
				Interface: elc.A.Interface,
			},
			Z: ExternalEndpoint{
				Node:      elc.Z.Node,
				External:  elc.Z.External,
				Network:   elc.Z.Network,
				Interface: elc.Z.Interface,
			},
		})
	}
}

// endpointRef returns a reference string for an external link endpoint.
func endpointRef(ep ExternalLinkEndpointConfig) string {
	if ep.Node != "" {
		return ep.Node
	}
	if ep.External != "" {
		return ep.External
	}
	return ep.Network
}

// applyAutoMgmt auto-generates management network display from node data.
func applyAutoMgmt(topo *Topology, cfg *Config, raw *RawTopology) {
	if cfg.AutoMgmt == nil || !cfg.AutoMgmt.Enabled {
		return
	}

	// Collect unique mgmt-net values from all nodes
	mgmtNets := map[string][]string{} // mgmt-net name → list of node names
	for _, node := range topo.Nodes {
		netName := node.MgmtNet
		if netName == "" {
			// Fallback to clab config
			if raw != nil {
				netName = ExtractClabMgmtNetwork(raw)
			}
			if netName == "" {
				netName = "clab"
			}
		}
		mgmtNets[netName] = append(mgmtNets[netName], node.Name)
	}

	// Sort network names for deterministic output
	netNames := make([]string, 0, len(mgmtNets))
	for name := range mgmtNets {
		netNames = append(netNames, name)
	}
	sort.Strings(netNames)

	for _, netName := range netNames {
		nodes := mgmtNets[netName]
		sort.Strings(nodes)

		networkID := fmt.Sprintf("mgmt:%s", netName)
		label := fmt.Sprintf("%s (mgmt)", netName)

		if cfg.AutoMgmt.Collapsed {
			// Collapsed mode: single network element with link count
			topo.ExternalNetworks = append(topo.ExternalNetworks, ExternalNetwork{
				Name:      networkID,
				Label:     label,
				Position:  cfg.AutoMgmt.Position,
				Collapsed: true,
				LinkCount: len(nodes),
			})
		} else {
			// Expanded mode: network + individual links to each node
			topo.ExternalNetworks = append(topo.ExternalNetworks, ExternalNetwork{
				Name:     networkID,
				Label:    label,
				Position: cfg.AutoMgmt.Position,
			})

			for _, nodeName := range nodes {
				id := fmt.Sprintf("ext:%s:mgmt__%s:mgmt", nodeName, networkID)
				topo.ExternalLinks = append(topo.ExternalLinks, ExternalLink{
					ID: id,
					A: ExternalEndpoint{
						Node:      nodeName,
						Interface: "mgmt",
					},
					Z: ExternalEndpoint{
						Network:   networkID,
						Interface: "mgmt",
					},
				})
			}
		}
	}
}
