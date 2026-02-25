package topology

import (
	"os"
	"testing"
)

func readTestData(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %v", name, err)
	}
	return data
}

func TestParseV073Format(t *testing.T) {
	data := readTestData(t, "topology-v073.json")
	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if topo.Name != "dc-fabric" {
		t.Errorf("expected name dc-fabric, got %s", topo.Name)
	}

	if len(topo.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(topo.Nodes))
	}

	if len(topo.Links) != 2 {
		t.Errorf("expected 2 links, got %d", len(topo.Links))
	}

	// Check link endpoints parsed from v0.73+ format
	for _, link := range topo.Links {
		if link.A.Node == "" || link.Z.Node == "" {
			t.Errorf("link %s has empty node names", link.ID)
		}
		if link.A.Interface == "" || link.Z.Interface == "" {
			t.Errorf("link %s has empty interface names", link.ID)
		}
	}
}

func TestParseLegacyFormat(t *testing.T) {
	data := readTestData(t, "topology-legacy.json")
	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if topo.Name != "legacy-lab" {
		t.Errorf("expected name legacy-lab, got %s", topo.Name)
	}

	if len(topo.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(topo.Nodes))
	}

	if len(topo.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(topo.Links))
	}

	// Check legacy link format
	link := topo.Links[0]
	if link.A.Node != "router1" {
		t.Errorf("expected link A node router1, got %s", link.A.Node)
	}
	if link.Z.Node != "switch1" {
		t.Errorf("expected link Z node switch1, got %s", link.Z.Node)
	}
}

func TestParseMinimal(t *testing.T) {
	data := readTestData(t, "topology-minimal.json")
	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if topo.Name != "minimal" {
		t.Errorf("expected name minimal, got %s", topo.Name)
	}

	if len(topo.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(topo.Nodes))
	}

	if len(topo.Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(topo.Links))
	}
}

func TestGraphHideFilter(t *testing.T) {
	data := []byte(`{
		"name": "hide-test",
		"nodes": {
			"visible": {
				"shortname": "visible",
				"kind": "linux",
				"labels": {}
			},
			"hidden": {
				"shortname": "hidden",
				"kind": "linux",
				"labels": {"graph-hide": "yes"}
			}
		},
		"links": []
	}`)

	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(topo.Nodes) != 1 {
		t.Fatalf("expected 1 visible node, got %d", len(topo.Nodes))
	}
	if topo.Nodes[0].Name != "visible" {
		t.Errorf("expected visible node, got %s", topo.Nodes[0].Name)
	}
}

func TestGrouping(t *testing.T) {
	data := readTestData(t, "topology-with-groups.json")
	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(topo.Groups.DCs) != 1 {
		t.Errorf("expected 1 DC, got %d", len(topo.Groups.DCs))
	}

	if topo.Groups.DCs[0] != "dc1" {
		t.Errorf("expected DC dc1, got %s", topo.Groups.DCs[0])
	}

	racks := topo.Groups.Racks["dc1"]
	if len(racks) != 2 {
		t.Errorf("expected 2 racks in dc1, got %d", len(racks))
	}
}

func TestIconResolution(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		labels   map[string]string
		bmc      bool
		wantIcon string
	}{
		{"srlinux", "nokia_srlinux", map[string]string{}, false, "switch"},
		{"ceos", "ceos", map[string]string{}, false, "switch"},
		{"crpd", "crpd", map[string]string{}, false, "router"},
		{"vr-router", "vr-sros", map[string]string{}, false, "router"},
		{"linux-server", "linux", map[string]string{"graph-role": "server"}, false, "server"},
		{"linux-bmc-label", "linux", map[string]string{"graph-bmc": "true"}, true, "bmc"},
		{"linux-bmc-image", "linux", map[string]string{}, true, "bmc"},
		{"linux-spine", "linux", map[string]string{"graph-role": "spine"}, false, "switch"},
		{"linux-default", "linux", map[string]string{}, false, "host"},
		{"explicit-icon", "linux", map[string]string{"graph-icon": "router"}, false, "router"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveIcon(tt.kind, tt.labels, tt.bmc)
			if got != tt.wantIcon {
				t.Errorf("resolveIcon(%q, %v, %v) = %q, want %q", tt.kind, tt.labels, tt.bmc, got, tt.wantIcon)
			}
		})
	}
}

func TestBMCAccessMethods(t *testing.T) {
	data := readTestData(t, "topology-with-bmc.json")
	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	for _, node := range topo.Nodes {
		hasVNC := false
		for _, am := range node.AccessMethods {
			if am.Type == "vnc" {
				hasVNC = true
				break
			}
		}
		if node.Labels["graph-bmc"] == "true" && !hasVNC {
			t.Errorf("node %s has graph-bmc=true but no VNC access method", node.Name)
		}
	}
}

func TestIsBMC(t *testing.T) {
	tests := []struct {
		name   string
		image  string
		labels map[string]string
		want   bool
	}{
		{"label-only", "alpine:latest", map[string]string{"graph-bmc": "true"}, true},
		{"image-only", "qemu-bmc:latest", map[string]string{}, true},
		{"image-with-registry", "ghcr.io/tjst-t/qemu-bmc:v1.0", map[string]string{}, true},
		{"image-case-insensitive", "Qemu-BMC:latest", map[string]string{}, true},
		{"neither", "alpine:latest", map[string]string{}, false},
		{"similar-name", "not-qemu-bmcd:latest", map[string]string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBMC(tt.image, tt.labels)
			if got != tt.want {
				t.Errorf("isBMC(%q, %v) = %v, want %v", tt.image, tt.labels, got, tt.want)
			}
		})
	}
}

func TestQemuBMCImageAutoDetectsVNC(t *testing.T) {
	// Test that a node with qemu-bmc image (without graph-bmc label) gets VNC access method
	data := []byte(`{
		"name": "auto-bmc",
		"nodes": {
			"bmc1": {
				"shortname": "bmc1",
				"kind": "linux",
				"image": "qemu-bmc:latest",
				"mgmt-ipv4-address": "172.20.20.10",
				"labels": {
					"containerlab": "auto-bmc"
				}
			},
			"server1": {
				"shortname": "server1",
				"kind": "linux",
				"image": "alpine:latest",
				"mgmt-ipv4-address": "172.20.20.11",
				"labels": {
					"containerlab": "auto-bmc"
				}
			}
		},
		"links": []
	}`)

	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	for _, node := range topo.Nodes {
		hasVNC := false
		var vncTarget string
		for _, am := range node.AccessMethods {
			if am.Type == "vnc" {
				hasVNC = true
				vncTarget = am.Target
			}
		}

		switch node.Name {
		case "bmc1":
			if !hasVNC {
				t.Errorf("node bmc1 (qemu-bmc image) should have VNC access method")
			}
			expectedTarget := "https://172.20.20.10/novnc/vnc.html"
			if vncTarget != expectedTarget {
				t.Errorf("bmc1 VNC target = %q, want %q", vncTarget, expectedTarget)
			}
			if node.Graph.Icon != "bmc" {
				t.Errorf("bmc1 icon = %q, want %q", node.Graph.Icon, "bmc")
			}
			if node.Graph.Role != "bmc" {
				t.Errorf("bmc1 role = %q, want %q", node.Graph.Role, "bmc")
			}
		case "server1":
			if hasVNC {
				t.Errorf("node server1 (alpine image) should NOT have VNC access method")
			}
		}
	}
}

func TestLinkID(t *testing.T) {
	data := readTestData(t, "topology-v073.json")
	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedID := "spine1:e1-1__leaf1:e1-49"
	found := false
	for _, link := range topo.Links {
		if link.ID == expectedID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected link ID %s not found", expectedID)
	}
}

func TestInvalidJSON(t *testing.T) {
	_, err := Parse([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNilLabels(t *testing.T) {
	data := []byte(`{
		"name": "nil-labels",
		"nodes": {
			"node1": {
				"shortname": "node1",
				"kind": "linux"
			}
		},
		"links": []
	}`)

	topo, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(topo.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(topo.Nodes))
	}

	if topo.Nodes[0].Labels == nil {
		t.Error("expected non-nil labels map")
	}
}
