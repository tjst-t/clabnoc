package topology

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseV073Format(t *testing.T) {
	topo, err := ParseFile("testdata/topology-v073.json")
	require.NoError(t, err)

	assert.Equal(t, "test-topo", topo.Name)
	assert.Len(t, topo.Nodes, 2)
	assert.Len(t, topo.Links, 1)

	// Nodes should be sorted by name
	assert.Equal(t, "leaf1", topo.Nodes[0].Name)
	assert.Equal(t, "spine1", topo.Nodes[1].Name)

	// Check node fields
	spine := topo.Nodes[1]
	assert.Equal(t, "spine1", spine.Name)
	assert.Equal(t, "nokia_srlinux", spine.Kind)
	assert.Equal(t, "172.20.20.2", spine.MgmtIPv4)
	assert.Equal(t, "dc1", spine.Graph.DC)
	assert.Equal(t, "rack1", spine.Graph.Rack)
	assert.Equal(t, "spine", spine.Graph.Role)
	assert.Equal(t, "switch", spine.Graph.Icon) // spine role -> switch icon

	leaf := topo.Nodes[0]
	assert.Equal(t, "leaf1", leaf.Name)
	assert.Equal(t, "172.20.20.3", leaf.MgmtIPv4)

	// Check link
	link := topo.Links[0]
	assert.Equal(t, "spine1:e1-1__leaf1:e1-49", link.ID)
	assert.Equal(t, "spine1", link.A.Node)
	assert.Equal(t, "e1-1", link.A.Interface)
	assert.Equal(t, "aa:c1:ab:01:01:01", link.A.MAC)
	assert.Equal(t, "leaf1", link.Z.Node)
	assert.Equal(t, "e1-49", link.Z.Interface)
	assert.Equal(t, "aa:c1:ab:01:01:02", link.Z.MAC)
	assert.Equal(t, "up", link.State)
}

func TestParseLegacyFormat(t *testing.T) {
	topo, err := ParseFile("testdata/topology-legacy.json")
	require.NoError(t, err)

	assert.Equal(t, "test-legacy", topo.Name)
	assert.Len(t, topo.Nodes, 2)
	assert.Len(t, topo.Links, 1)

	// Check legacy link was parsed correctly
	link := topo.Links[0]
	assert.Equal(t, "spine1:e1-1__leaf1:e1-49", link.ID)
	assert.Equal(t, "spine1", link.A.Node)
	assert.Equal(t, "e1-1", link.A.Interface)
	assert.Equal(t, "aa:c1:ab:01:01:01", link.A.MAC)
	assert.Equal(t, "leaf1", link.Z.Node)
	assert.Equal(t, "e1-49", link.Z.Interface)
	assert.Equal(t, "aa:c1:ab:01:01:02", link.Z.MAC)
	assert.Equal(t, "up", link.State)
}

func TestParseMinimalTopology(t *testing.T) {
	topo, err := ParseFile("testdata/topology-minimal.json")
	require.NoError(t, err)

	assert.Equal(t, "minimal", topo.Name)
	assert.Len(t, topo.Nodes, 1)
	assert.Len(t, topo.Links, 0)

	node := topo.Nodes[0]
	assert.Equal(t, "host1", node.Name)
	assert.Equal(t, "linux", node.Kind)
	assert.Equal(t, "172.20.20.2", node.MgmtIPv4)
}

func TestGraphHideLabel(t *testing.T) {
	topo, err := ParseFile("testdata/topology-with-groups.json")
	require.NoError(t, err)

	// Find server1 which has graph-hide: yes
	var server1 *Node
	for i := range topo.Nodes {
		if topo.Nodes[i].Name == "server1" {
			server1 = &topo.Nodes[i]
			break
		}
	}
	require.NotNil(t, server1)
	assert.True(t, server1.Graph.Hidden)

	// Other nodes should not be hidden
	for _, n := range topo.Nodes {
		if n.Name != "server1" {
			assert.False(t, n.Graph.Hidden, "node %s should not be hidden", n.Name)
		}
	}
}

func TestDCRackGrouping(t *testing.T) {
	topo, err := ParseFile("testdata/topology-with-groups.json")
	require.NoError(t, err)

	// Should have dc1 and dc2
	assert.Contains(t, topo.Groups.DCs, "dc1")
	assert.Contains(t, topo.Groups.DCs, "dc2")
	assert.Len(t, topo.Groups.DCs, 2)

	// dc1 racks: spine-rack, rack1, rack2
	dc1Racks := topo.Groups.Racks["dc1"]
	assert.Contains(t, dc1Racks, "rack1")
	assert.Contains(t, dc1Racks, "rack2")
	assert.Contains(t, dc1Racks, "spine-rack")
	assert.Len(t, dc1Racks, 3)

	// dc2 racks: rack1
	dc2Racks := topo.Groups.Racks["dc2"]
	assert.Contains(t, dc2Racks, "rack1")
	assert.Len(t, dc2Racks, 1)
}

func TestBMCNodeAccessMethods(t *testing.T) {
	topo, err := ParseFile("testdata/topology-with-bmc.json")
	require.NoError(t, err)

	// Find bmc1
	var bmc1 *Node
	for i := range topo.Nodes {
		if topo.Nodes[i].Name == "bmc1" {
			bmc1 = &topo.Nodes[i]
			break
		}
	}
	require.NotNil(t, bmc1)

	// BMC should have exec, ssh, and vnc access methods
	methodTypes := map[string]bool{}
	for _, am := range bmc1.AccessMethods {
		methodTypes[am.Type] = true
	}
	assert.True(t, methodTypes["exec"], "exec method should be present")
	assert.True(t, methodTypes["vnc"], "vnc method should be present for BMC node")

	// Check VNC target
	for _, am := range bmc1.AccessMethods {
		if am.Type == "vnc" {
			assert.Equal(t, "https://172.20.20.3/novnc/vnc.html", am.Target)
		}
	}

	// BMC icon should be "bmc"
	assert.Equal(t, "bmc", bmc1.Graph.Icon)
}

func TestAccessMethodsForSRLinux(t *testing.T) {
	topo, err := ParseFile("testdata/topology-v073.json")
	require.NoError(t, err)

	spine := topo.Nodes[1]
	assert.Equal(t, "spine1", spine.Name)

	// Should have exec and ssh
	methodTypes := map[string]bool{}
	for _, am := range spine.AccessMethods {
		methodTypes[am.Type] = true
	}
	assert.True(t, methodTypes["exec"], "exec should always be present")
	assert.True(t, methodTypes["ssh"], "ssh should be present for nokia_srlinux")
	assert.False(t, methodTypes["vnc"], "vnc should not be present for non-BMC node")

	// Check SSH target
	for _, am := range spine.AccessMethods {
		if am.Type == "ssh" {
			assert.Equal(t, "172.20.20.2:22", am.Target)
		}
	}
}

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantName  string
		wantNodes int
		wantLinks int
		wantErr   bool
	}{
		{
			name:    "invalid json",
			json:    `{not valid json}`,
			wantErr: true,
		},
		{
			name: "empty topology",
			json: `{"name": "empty", "nodes": {}, "links": []}`,
			wantName:  "empty",
			wantNodes: 0,
			wantLinks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topo, err := ParseBytes([]byte(tt.json))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, topo.Name)
			assert.Len(t, topo.Nodes, tt.wantNodes)
			assert.Len(t, topo.Links, tt.wantLinks)
		})
	}
}

func TestInferIcon(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		role     string
		labels   map[string]string
		wantIcon string
	}{
		{"nokia_srlinux", "nokia_srlinux", "", map[string]string{}, "switch"},
		{"ceos", "ceos", "", map[string]string{}, "switch"},
		{"crpd", "crpd", "", map[string]string{}, "router"},
		{"spine role", "", "spine", map[string]string{}, "switch"},
		{"leaf role", "", "leaf", map[string]string{}, "switch"},
		{"server role", "", "server", map[string]string{}, "server"},
		{"bmc role", "", "bmc", map[string]string{}, "bmc"},
		{"bmc label", "linux", "", map[string]string{"graph-bmc": "true"}, "bmc"},
		{"default host", "linux", "", map[string]string{}, "host"},
		{"unknown kind", "unknown-kind", "", map[string]string{}, "host"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := inferIcon(tt.kind, tt.role, tt.labels)
			assert.Equal(t, tt.wantIcon, icon)
		})
	}
}
