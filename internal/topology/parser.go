package topology

import (
	"encoding/json"
	"fmt"
	"os"
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

	for name, rn := range raw.Nodes {
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

	graph := GraphInfo{
		DC:     labels["graph-dc"],
		Rack:   labels["graph-rack"],
		Role:   labels["graph-role"],
		Icon:   resolveIcon(rn.Kind, labels),
		Hidden: labels["graph-hide"] == "yes",
	}

	if u, err := strconv.Atoi(labels["graph-rack-unit"]); err == nil {
		graph.RackUnit = u
	}

	if graph.Role == "" {
		graph.Role = inferRole(rn.Kind, labels)
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
	if labels["graph-bmc"] == "true" && rn.MgmtIPv4 != "" {
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

func resolveIcon(kind string, labels map[string]string) string {
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
		if role == "bmc" || labels["graph-bmc"] == "true" {
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

func inferRole(kind string, labels map[string]string) string {
	kind = strings.ToLower(kind)
	switch {
	case kind == "nokia_srlinux" || kind == "ceos":
		return "switch"
	case kind == "crpd" || strings.HasPrefix(kind, "vr-"):
		return "router"
	case kind == "linux":
		if labels["graph-bmc"] == "true" {
			return "bmc"
		}
		return "server"
	default:
		return ""
	}
}
