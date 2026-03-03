package topology

import (
	"testing"
)

func TestApplyExternalNodes(t *testing.T) {
	topo := &Topology{
		Name:  "test",
		Nodes: []Node{},
		Groups: Groups{
			DCs:   []string{"dc-a"},
			Racks: map[string][]string{"dc-a": {"rack-a01"}},
		},
	}

	cfg := &Config{
		Racks: map[string]RackConfig{
			"rack-a01": {DC: "dc-a", Units: 42},
		},
		ExternalNodes: map[string]ExternalNodeConfig{
			"ntp-server": {
				Label:       "NTP Server",
				Description: "Campus NTP",
				Icon:        "service",
				Interfaces:  []string{"eth0"},
				Placement:   ExternalNodePlacement{DC: "dc-a"},
			},
			"oob-switch": {
				Label:      "OOB Switch",
				Icon:       "switch",
				Interfaces: []string{"ge-0/0/0", "ge-0/0/1"},
				Placement:  ExternalNodePlacement{DC: "dc-a", Rack: "rack-a01", RackUnit: 20, Size: 1},
			},
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNodes) != 2 {
		t.Fatalf("expected 2 external nodes, got %d", len(topo.ExternalNodes))
	}

	// Sorted alphabetically: ntp-server, oob-switch
	ntp := topo.ExternalNodes[0]
	if ntp.Name != "ntp-server" {
		t.Errorf("expected ntp-server, got %s", ntp.Name)
	}
	if ntp.Label != "NTP Server" {
		t.Errorf("label = %q, want %q", ntp.Label, "NTP Server")
	}
	if !ntp.External {
		t.Error("expected external=true")
	}
	if ntp.Graph.DC != "dc-a" {
		t.Errorf("DC = %q, want %q", ntp.Graph.DC, "dc-a")
	}

	oob := topo.ExternalNodes[1]
	if oob.Name != "oob-switch" {
		t.Errorf("expected oob-switch, got %s", oob.Name)
	}
	if oob.Graph.Rack != "rack-a01" {
		t.Errorf("rack = %q, want %q", oob.Graph.Rack, "rack-a01")
	}
	if oob.Graph.RackUnit != 20 {
		t.Errorf("rack_unit = %d, want %d", oob.Graph.RackUnit, 20)
	}
}

func TestApplyExternalNodesDCFromRack(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Groups: Groups{
			DCs:   []string{"dc-a"},
			Racks: map[string][]string{"dc-a": {"rack-a01"}},
		},
	}

	cfg := &Config{
		Racks: map[string]RackConfig{
			"rack-a01": {DC: "dc-a", Units: 42},
		},
		ExternalNodes: map[string]ExternalNodeConfig{
			"node-in-rack": {
				Label: "Rack Node",
				Icon:  "service",
				Placement: ExternalNodePlacement{
					Rack:     "rack-a01",
					RackUnit: 5,
					Size:     1,
				},
			},
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNodes) != 1 {
		t.Fatalf("expected 1 external node, got %d", len(topo.ExternalNodes))
	}

	node := topo.ExternalNodes[0]
	if node.Graph.DC != "dc-a" {
		t.Errorf("DC should be resolved from rack, got %q", node.Graph.DC)
	}
}

func TestApplyExternalNetworks(t *testing.T) {
	topo := &Topology{Name: "test"}

	cfg := &Config{
		ExternalNetworks: map[string]ExternalNetworkConfig{
			"internet": {Label: "Internet", Position: "top"},
			"oob":      {Label: "OOB Management", Position: "bottom", DC: "dc-a"},
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNetworks) != 2 {
		t.Fatalf("expected 2 external networks, got %d", len(topo.ExternalNetworks))
	}

	// Sorted alphabetically: internet, oob
	inet := topo.ExternalNetworks[0]
	if inet.Name != "internet" {
		t.Errorf("expected internet, got %s", inet.Name)
	}
	if inet.Position != "top" {
		t.Errorf("position = %q, want %q", inet.Position, "top")
	}

	oob := topo.ExternalNetworks[1]
	if oob.DC != "dc-a" {
		t.Errorf("dc = %q, want %q", oob.DC, "dc-a")
	}
}

func TestApplyExternalLinks(t *testing.T) {
	topo := &Topology{Name: "test"}

	cfg := &Config{
		ExternalLinks: []ExternalLinkConfig{
			{
				A: ExternalLinkEndpointConfig{Node: "spine1", Interface: "e1-48"},
				Z: ExternalLinkEndpointConfig{Network: "internet"},
			},
			{
				A: ExternalLinkEndpointConfig{External: "oob-switch", Interface: "ge-0/0/0"},
				Z: ExternalLinkEndpointConfig{Network: "oob-mgmt"},
			},
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalLinks) != 2 {
		t.Fatalf("expected 2 external links, got %d", len(topo.ExternalLinks))
	}

	link1 := topo.ExternalLinks[0]
	if link1.ID != "ext:spine1:e1-48__internet:" {
		t.Errorf("link1 ID = %q, want %q", link1.ID, "ext:spine1:e1-48__internet:")
	}
	if link1.A.Node != "spine1" {
		t.Errorf("link1 A.Node = %q", link1.A.Node)
	}
	if link1.Z.Network != "internet" {
		t.Errorf("link1 Z.Network = %q", link1.Z.Network)
	}

	link2 := topo.ExternalLinks[1]
	if link2.A.External != "oob-switch" {
		t.Errorf("link2 A.External = %q", link2.A.External)
	}
}

func TestApplyAutoMgmtCollapsed(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "spine1", MgmtNet: "clab"},
			{Name: "leaf1", MgmtNet: "clab"},
			{Name: "server1", MgmtNet: "clab"},
		},
	}

	cfg := &Config{
		AutoMgmt: &AutoMgmtConfig{
			Enabled:   true,
			Position:  "bottom",
			Collapsed: true,
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNetworks) != 1 {
		t.Fatalf("expected 1 external network, got %d", len(topo.ExternalNetworks))
	}

	net := topo.ExternalNetworks[0]
	if net.Name != "mgmt:clab" {
		t.Errorf("name = %q, want %q", net.Name, "mgmt:clab")
	}
	if net.Label != "clab (mgmt)" {
		t.Errorf("label = %q, want %q", net.Label, "clab (mgmt)")
	}
	if !net.Collapsed {
		t.Error("expected collapsed=true")
	}
	if net.LinkCount != 3 {
		t.Errorf("link_count = %d, want %d", net.LinkCount, 3)
	}
	if net.Position != "bottom" {
		t.Errorf("position = %q, want %q", net.Position, "bottom")
	}

	// No individual links should be generated in collapsed mode
	if len(topo.ExternalLinks) != 0 {
		t.Errorf("expected 0 external links in collapsed mode, got %d", len(topo.ExternalLinks))
	}
}

func TestApplyAutoMgmtExpanded(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "spine1", MgmtNet: "clab"},
			{Name: "leaf1", MgmtNet: "clab"},
		},
	}

	cfg := &Config{
		AutoMgmt: &AutoMgmtConfig{
			Enabled:   true,
			Position:  "bottom",
			Collapsed: false,
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNetworks) != 1 {
		t.Fatalf("expected 1 external network, got %d", len(topo.ExternalNetworks))
	}

	net := topo.ExternalNetworks[0]
	if net.Collapsed {
		t.Error("expected collapsed=false")
	}

	// Should generate individual links
	if len(topo.ExternalLinks) != 2 {
		t.Fatalf("expected 2 external links in expanded mode, got %d", len(topo.ExternalLinks))
	}

	// Links sorted by node name
	link1 := topo.ExternalLinks[0]
	if link1.A.Node != "leaf1" {
		t.Errorf("link1 A.Node = %q, want %q", link1.A.Node, "leaf1")
	}
	if link1.Z.Network != "mgmt:clab" {
		t.Errorf("link1 Z.Network = %q, want %q", link1.Z.Network, "mgmt:clab")
	}
}

func TestApplyAutoMgmtFallback(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "spine1"}, // No MgmtNet set
		},
	}

	raw := &RawTopology{
		Clab: map[string]interface{}{
			"config": map[string]interface{}{
				"mgmt": map[string]interface{}{
					"network": "custom-mgmt",
				},
			},
		},
	}

	cfg := &Config{
		AutoMgmt: &AutoMgmtConfig{
			Enabled:   true,
			Position:  "bottom",
			Collapsed: true,
		},
	}

	ApplyExternalConfig(topo, cfg, raw)

	if len(topo.ExternalNetworks) != 1 {
		t.Fatalf("expected 1 external network, got %d", len(topo.ExternalNetworks))
	}

	net := topo.ExternalNetworks[0]
	if net.Name != "mgmt:custom-mgmt" {
		t.Errorf("name = %q, want %q", net.Name, "mgmt:custom-mgmt")
	}
}

func TestApplyAutoMgmtDefaultFallback(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "spine1"}, // No MgmtNet, no clab config
		},
	}

	cfg := &Config{
		AutoMgmt: &AutoMgmtConfig{
			Enabled:   true,
			Position:  "bottom",
			Collapsed: true,
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNetworks) != 1 {
		t.Fatalf("expected 1 external network, got %d", len(topo.ExternalNetworks))
	}

	net := topo.ExternalNetworks[0]
	if net.Name != "mgmt:clab" {
		t.Errorf("name = %q, want %q (default fallback)", net.Name, "mgmt:clab")
	}
}

func TestApplyAutoMgmtDisabled(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "spine1", MgmtNet: "clab"},
		},
	}

	cfg := &Config{
		AutoMgmt: &AutoMgmtConfig{
			Enabled: false,
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNetworks) != 0 {
		t.Errorf("expected 0 external networks when disabled, got %d", len(topo.ExternalNetworks))
	}
}

func TestApplyAutoMgmtMultipleNets(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "spine1", MgmtNet: "mgmt-a"},
			{Name: "leaf1", MgmtNet: "mgmt-b"},
			{Name: "server1", MgmtNet: "mgmt-a"},
		},
	}

	cfg := &Config{
		AutoMgmt: &AutoMgmtConfig{
			Enabled:   true,
			Position:  "bottom",
			Collapsed: true,
		},
	}

	ApplyExternalConfig(topo, cfg, nil)

	if len(topo.ExternalNetworks) != 2 {
		t.Fatalf("expected 2 external networks, got %d", len(topo.ExternalNetworks))
	}

	// Sorted: mgmt-a, mgmt-b
	netA := topo.ExternalNetworks[0]
	if netA.Name != "mgmt:mgmt-a" {
		t.Errorf("net A name = %q", netA.Name)
	}
	if netA.LinkCount != 2 {
		t.Errorf("net A link_count = %d, want 2", netA.LinkCount)
	}

	netB := topo.ExternalNetworks[1]
	if netB.Name != "mgmt:mgmt-b" {
		t.Errorf("net B name = %q", netB.Name)
	}
	if netB.LinkCount != 1 {
		t.Errorf("net B link_count = %d, want 1", netB.LinkCount)
	}
}

func TestApplyExternalConfigNil(t *testing.T) {
	topo := &Topology{Name: "test"}

	// Should not panic
	ApplyExternalConfig(topo, nil, nil)

	if len(topo.ExternalNodes) != 0 {
		t.Error("expected no external nodes")
	}
}

func TestExtractClabMgmtNetwork(t *testing.T) {
	tests := []struct {
		name string
		raw  *RawTopology
		want string
	}{
		{
			"with network",
			&RawTopology{
				Clab: map[string]interface{}{
					"config": map[string]interface{}{
						"mgmt": map[string]interface{}{
							"network": "clab",
						},
					},
				},
			},
			"clab",
		},
		{
			"no clab",
			&RawTopology{},
			"",
		},
		{
			"no config",
			&RawTopology{Clab: map[string]interface{}{}},
			"",
		},
		{
			"no mgmt",
			&RawTopology{
				Clab: map[string]interface{}{
					"config": map[string]interface{}{},
				},
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractClabMgmtNetwork(tt.raw)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
