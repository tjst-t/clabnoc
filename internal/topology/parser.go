package topology

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
)

// ParseFile reads and parses topology-data.json
func ParseFile(path string) (*Topology, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read topology file: %w", err)
	}
	return ParseBytes(data)
}

// ParseBytes parses topology-data.json from bytes
func ParseBytes(data []byte) (*Topology, error) {
	var raw topoDataRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal topology: %w", err)
	}

	topo := &Topology{
		Name:  raw.Name,
		Nodes: []Node{}, // initialize to empty slice so JSON serializes as [] not null
		Links: []Link{}, // initialize to empty slice so JSON serializes as [] not null
	}

	// Parse nodes (map -> slice, sorted by name for determinism)
	for shortName, n := range raw.Nodes {
		node := parseNode(shortName, n)
		topo.Nodes = append(topo.Nodes, node)
	}
	sort.Slice(topo.Nodes, func(i, j int) bool { return topo.Nodes[i].Name < topo.Nodes[j].Name })

	// Parse links
	for _, rawLink := range raw.Links {
		link, err := parseLink(rawLink)
		if err != nil {
			return nil, fmt.Errorf("parse link: %w", err)
		}
		link.ID = link.A.Node + ":" + link.A.Interface + "__" + link.Z.Node + ":" + link.Z.Interface
		link.State = "up" // default
		topo.Links = append(topo.Links, link)
	}

	// Build groups from labels
	topo.Groups = buildGroups(topo.Nodes)

	return topo, nil
}

// parseNode converts nodeRaw to Node including graph info from labels
func parseNode(shortName string, n nodeRaw) Node {
	node := Node{
		Name:      shortName,
		ShortName: n.ShortName,
		LongName:  n.LongName,
		FQDN:      n.FQDN,
		Kind:      n.Kind,
		Image:     n.Image,
		MgmtIPv4:  n.MgmtIPv4Address,
		MgmtIPv6:  n.MgmtIPv6Address,
		Labels:    n.Labels,
		Status:    "unknown",
	}

	// Use shortName as Name if ShortName is not set
	if node.ShortName == "" {
		node.ShortName = shortName
	}

	// Graph info from labels
	node.Graph = GraphInfo{
		DC:     n.Labels["graph-dc"],
		Rack:   n.Labels["graph-rack"],
		Role:   n.Labels["graph-role"],
		Icon:   n.Labels["graph-icon"],
		Hidden: n.Labels["graph-hide"] == "yes",
	}
	if ru, ok := n.Labels["graph-rack-unit"]; ok && ru != "" {
		if v, err := strconv.Atoi(ru); err == nil {
			node.Graph.RackUnit = v
		}
	}
	if node.Graph.Icon == "" {
		node.Graph.Icon = inferIcon(n.Kind, node.Graph.Role, n.Labels)
	}

	// Port bindings
	for _, pb := range n.PortBindings {
		node.PortBindings = append(node.PortBindings, PortBinding{
			HostIP:   pb.HostIP,
			HostPort: pb.HostPort,
			Port:     pb.Port,
			Protocol: pb.Protocol,
		})
	}

	// Access methods
	node.AccessMethods = inferAccessMethods(n.Kind, node.Graph.Role, n.Labels, n.MgmtIPv4Address)

	return node
}

// inferIcon determines icon from kind, role, labels
func inferIcon(kind, role string, labels map[string]string) string {
	if labels["graph-bmc"] == "true" || role == "bmc" {
		return "bmc"
	}
	switch kind {
	case "nokia_srlinux", "ceos":
		return "switch"
	case "crpd", "vr-sros", "vr-xrv9k", "vr-vmx", "vr-csr":
		return "router"
	}
	switch role {
	case "spine", "leaf", "oob":
		return "switch"
	case "server":
		return "server"
	case "bmc":
		return "bmc"
	}
	return "host"
}

// inferAccessMethods determines available access methods
func inferAccessMethods(kind, role string, labels map[string]string, mgmtIP string) []AccessMethod {
	var methods []AccessMethod
	// exec always available
	methods = append(methods, AccessMethod{Type: "exec", Label: "Console (docker exec)"})
	// ssh for common kinds
	if kind != "" && kind != "bridge" {
		target := ""
		if mgmtIP != "" {
			target = mgmtIP + ":22"
		}
		methods = append(methods, AccessMethod{Type: "ssh", Label: "SSH", Target: target})
	}
	// vnc for bmc
	if labels["graph-bmc"] == "true" && mgmtIP != "" {
		methods = append(methods, AccessMethod{Type: "vnc", Label: "noVNC", Target: "https://" + mgmtIP + "/novnc/vnc.html"})
	}
	return methods
}

// buildGroups extracts DC/rack structure from node labels
func buildGroups(nodes []Node) Groups {
	dcSet := map[string]bool{}
	racksByDC := map[string]map[string]bool{}
	for _, n := range nodes {
		dc := n.Graph.DC
		rack := n.Graph.Rack
		if dc != "" {
			dcSet[dc] = true
			if rack != "" {
				if racksByDC[dc] == nil {
					racksByDC[dc] = map[string]bool{}
				}
				racksByDC[dc][rack] = true
			}
		}
	}
	g := Groups{
		DCs:   []string{},            // initialize to empty slice so JSON serializes as [] not null
		Racks: map[string][]string{}, // already non-nil
	}
	for dc := range dcSet {
		g.DCs = append(g.DCs, dc)
	}
	sort.Strings(g.DCs)
	for dc, racks := range racksByDC {
		for rack := range racks {
			g.Racks[dc] = append(g.Racks[dc], rack)
		}
		sort.Strings(g.Racks[dc])
	}
	return g
}
