package topology

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ParseFile reads and parses a topology-data.json file.
func ParseFile(path string) (*Topology, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading topology file: %w", err)
	}
	return Parse(data)
}

// Parse parses topology-data.json bytes into a Topology.
func Parse(data []byte) (*Topology, error) {
	var raw RawTopology
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing topology JSON: %w", err)
	}
	return Convert(&raw), nil
}

// Convert transforms a RawTopology into a processed Topology.
func Convert(raw *RawTopology) *Topology {
	topo := &Topology{
		Name:  raw.Name,
		Links: make([]Link, 0),
		Nodes: make([]Node, 0),
		Groups: Groups{
			DCs:   make([]string, 0),
			Racks: make(map[string][]string),
		},
	}

	dcSet := map[string]bool{}
	racksByDC := map[string]map[string]bool{}

	// Sort node names for deterministic output
	nodeNames := make([]string, 0, len(raw.Nodes))
	for name := range raw.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	for _, name := range nodeNames {
		rn := raw.Nodes[name]
		node := convertNode(name, rn)
		if node.Graph.Hidden {
			continue
		}
		topo.Nodes = append(topo.Nodes, node)

		if dc := node.Graph.DC; dc != "" {
			if !dcSet[dc] {
				dcSet[dc] = true
				topo.Groups.DCs = append(topo.Groups.DCs, dc)
				racksByDC[dc] = map[string]bool{}
			}
			if rack := node.Graph.Rack; rack != "" {
				if !racksByDC[dc][rack] {
					racksByDC[dc][rack] = true
					topo.Groups.Racks[dc] = append(topo.Groups.Racks[dc], rack)
				}
			}
		}
	}

	// Sort DCs alphabetically
	sort.Strings(topo.Groups.DCs)

	// Sort racks within each DC alphabetically
	for dc := range topo.Groups.Racks {
		sort.Strings(topo.Groups.Racks[dc])
	}

	for _, rl := range raw.Links {
		link := convertLink(rl)
		topo.Links = append(topo.Links, link)
	}

	return topo
}

func convertNode(name string, rn RawNode) Node {
	labels := rn.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	bmc := isBMC(rn.Image, labels)

	graph := GraphInfo{
		DC:     labels["graph-dc"],
		Rack:   labels["graph-rack"],
		Role:   labels["graph-role"],
		Icon:   resolveIcon(rn.Kind, labels, bmc),
		Hidden: labels["graph-hide"] == "yes",
	}

	if u, err := strconv.Atoi(labels["graph-rack-unit"]); err == nil {
		graph.RackUnit = u
	}

	if s, err := strconv.Atoi(labels["graph-rack-unit-size"]); err == nil && s > 0 {
		graph.RackUnitSize = s
	} else {
		graph.RackUnitSize = 1
	}

	if graph.Role == "" {
		graph.Role = inferRole(rn.Kind, labels, bmc)
	}

	accessMethods := []AccessMethod{
		{Type: "exec", Label: "Console (docker exec)"},
	}
	if rn.MgmtIPv4 != "" {
		accessMethods = append(accessMethods, AccessMethod{
			Type:   "ssh",
			Label:  "SSH",
			Target: rn.MgmtIPv4 + ":22",
		})
	}
	if bmc && rn.MgmtIPv4 != "" {
		accessMethods = append(accessMethods, AccessMethod{
			Type:   "vnc",
			Label:  "noVNC (BMC)",
			Target: fmt.Sprintf("https://%s/novnc/vnc.html", rn.MgmtIPv4),
		})
	}

	return Node{
		Name:          name,
		Kind:          rn.Kind,
		Image:         rn.Image,
		Status:        "running",
		MgmtIPv4:     rn.MgmtIPv4,
		MgmtIPv6:     rn.MgmtIPv6,
		MgmtNet:      rn.MgmtNet,
		Labels:        labels,
		PortBindings:  rn.PortBindings,
		AccessMethods: accessMethods,
		Graph:         graph,
	}
}

func convertLink(rl RawLink) Link {
	var a, z Endpoint

	if rl.Endpoints != nil {
		a = Endpoint{Node: rl.Endpoints.A.Node, Interface: rl.Endpoints.A.Interface, MAC: rl.Endpoints.A.MAC}
		z = Endpoint{Node: rl.Endpoints.Z.Node, Interface: rl.Endpoints.Z.Interface, MAC: rl.Endpoints.Z.MAC}
	} else if rl.A != nil && rl.Z != nil {
		a = Endpoint{Node: rl.A.Node, Interface: rl.A.Interface, MAC: rl.A.MAC}
		z = Endpoint{Node: rl.Z.Node, Interface: rl.Z.Interface, MAC: rl.Z.MAC}
	}

	id := fmt.Sprintf("%s:%s__%s:%s", a.Node, a.Interface, z.Node, z.Interface)

	return Link{
		ID:    id,
		A:     a,
		Z:     z,
		State: "up",
	}
}

// isBMC returns true if the node is a BMC, detected by label or image name.
func isBMC(image string, labels map[string]string) bool {
	if labels["graph-bmc"] == "true" {
		return true
	}
	// Auto-detect qemu-bmc by image name (e.g. "qemu-bmc:latest", "ghcr.io/foo/qemu-bmc:v1")
	return strings.Contains(strings.ToLower(image), "qemu-bmc")
}

func resolveIcon(kind string, labels map[string]string, bmc bool) string {
	if icon := labels["graph-icon"]; icon != "" {
		return icon
	}

	kind = strings.ToLower(kind)
	switch {
	case kind == "nokia_srlinux" || kind == "ceos":
		return "switch"
	case kind == "crpd" || strings.HasPrefix(kind, "vr-"):
		return "router"
	case kind == "linux":
		role := labels["graph-role"]
		if role == "bmc" || bmc {
			return "bmc"
		}
		if role == "spine" || role == "leaf" {
			return "switch"
		}
		if role == "server" {
			return "server"
		}
		return "host"
	default:
		return "host"
	}
}

// ExtractClabMgmtNetwork extracts the management network name from the clab config
// in topology-data.json. Returns empty string if not found.
func ExtractClabMgmtNetwork(raw *RawTopology) string {
	if raw.Clab == nil {
		return ""
	}
	config, ok := raw.Clab["config"].(map[string]interface{})
	if !ok {
		return ""
	}
	mgmt, ok := config["mgmt"].(map[string]interface{})
	if !ok {
		return ""
	}
	network, ok := mgmt["network"].(string)
	if !ok {
		return ""
	}
	return network
}

// ParseRaw parses topology-data.json bytes into a RawTopology.
func ParseRaw(data []byte) (*RawTopology, error) {
	var raw RawTopology
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing topology JSON: %w", err)
	}
	return &raw, nil
}

func inferRole(kind string, labels map[string]string, bmc bool) string {
	kind = strings.ToLower(kind)
	switch {
	case kind == "nokia_srlinux" || kind == "ceos":
		return "switch"
	case kind == "crpd" || strings.HasPrefix(kind, "vr-"):
		return "router"
	case kind == "linux":
		if bmc {
			return "bmc"
		}
		return "server"
	default:
		return ""
	}
}
