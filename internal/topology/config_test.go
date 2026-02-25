package topology

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	cfg, err := LoadConfigFile("../../testdata/sample.clabnoc.yml")
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Check racks
	if len(cfg.Racks) != 2 {
		t.Fatalf("expected 2 racks, got %d", len(cfg.Racks))
	}
	if cfg.Racks["rack-a01"].DC != "dc-a" {
		t.Errorf("rack-a01 DC = %q, want %q", cfg.Racks["rack-a01"].DC, "dc-a")
	}
	if cfg.Racks["rack-a01"].Units != 42 {
		t.Errorf("rack-a01 Units = %d, want %d", cfg.Racks["rack-a01"].Units, 42)
	}
	if cfg.Racks["rack-a02"].Units != 48 {
		t.Errorf("rack-a02 Units = %d, want %d", cfg.Racks["rack-a02"].Units, 48)
	}

	// Check nodes
	if len(cfg.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(cfg.Nodes))
	}
	compute := cfg.Nodes["compute-01"]
	if compute.Rack != "rack-a01" {
		t.Errorf("compute-01 Rack = %q, want %q", compute.Rack, "rack-a01")
	}
	if compute.Unit != 39 {
		t.Errorf("compute-01 Unit = %d, want %d", compute.Unit, 39)
	}
	if compute.Size != 2 {
		t.Errorf("compute-01 Size = %d, want %d", compute.Size, 2)
	}
	if compute.Role != "server" {
		t.Errorf("compute-01 Role = %q, want %q", compute.Role, "server")
	}

	gpu := cfg.Nodes["gpu-node-01"]
	if gpu.Size != 4 {
		t.Errorf("gpu-node-01 Size = %d, want %d", gpu.Size, 4)
	}

	// Check kind_defaults
	if len(cfg.KindDefaults) != 2 {
		t.Fatalf("expected 2 kind_defaults, got %d", len(cfg.KindDefaults))
	}
	srlKD := cfg.KindDefaults["nokia_srlinux"]
	if srlKD.SSH == nil {
		t.Fatal("nokia_srlinux kind_defaults SSH is nil")
	}
	if srlKD.SSH.Username != "admin" {
		t.Errorf("nokia_srlinux SSH Username = %q, want %q", srlKD.SSH.Username, "admin")
	}
	if srlKD.SSH.Password != "NokiaSrl1!" {
		t.Errorf("nokia_srlinux SSH Password = %q, want %q", srlKD.SSH.Password, "NokiaSrl1!")
	}

	// Check node-level SSH
	if compute.SSH == nil {
		t.Fatal("compute-01 SSH is nil")
	}
	if compute.SSH.Username != "ubuntu" {
		t.Errorf("compute-01 SSH Username = %q, want %q", compute.SSH.Username, "ubuntu")
	}
	if compute.SSH.Port != 2222 {
		t.Errorf("compute-01 SSH Port = %d, want %d", compute.SSH.Port, 2222)
	}
}

func TestLoadConfigFileDefaults(t *testing.T) {
	// Test that default values are applied
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.clabnoc.yml")

	content := []byte(`
racks:
  rack-01:
    dc: dc1
nodes:
  node1:
    rack: rack-01
    unit: 10
`)
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// Rack units should default to 42
	if cfg.Racks["rack-01"].Units != 42 {
		t.Errorf("rack-01 Units default = %d, want 42", cfg.Racks["rack-01"].Units)
	}

	// Node size should default to 1
	if cfg.Nodes["node1"].Size != 1 {
		t.Errorf("node1 Size default = %d, want 1", cfg.Nodes["node1"].Size)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfigFile("/nonexistent/path.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "invalid.yml")
	if err := os.WriteFile(cfgPath, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfigFile(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Simulate labDir = /tmp/xxx/clab-mylab
	labDir := filepath.Join(tmpDir, "clab-mylab")
	if err := os.MkdirAll(labDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("parent directory config", func(t *testing.T) {
		// Create <labDir>/../mylab.clabnoc.yml
		cfgPath := filepath.Join(tmpDir, "mylab.clabnoc.yml")
		if err := os.WriteFile(cfgPath, []byte("racks: {}"), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(cfgPath)

		result := FindConfigFile(labDir, "mylab")
		if result == "" {
			t.Error("expected to find config in parent directory")
		}
	})

	t.Run("labdir config", func(t *testing.T) {
		// Create <labDir>/clabnoc.yml
		cfgPath := filepath.Join(labDir, "clabnoc.yml")
		if err := os.WriteFile(cfgPath, []byte("racks: {}"), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(cfgPath)

		result := FindConfigFile(labDir, "mylab")
		if result == "" {
			t.Error("expected to find config in labdir")
		}
	})

	t.Run("parent takes priority", func(t *testing.T) {
		// Create both files
		parentCfg := filepath.Join(tmpDir, "mylab.clabnoc.yml")
		labdirCfg := filepath.Join(labDir, "clabnoc.yml")
		if err := os.WriteFile(parentCfg, []byte("racks: {}"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(labdirCfg, []byte("racks: {}"), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(parentCfg)
		defer os.Remove(labdirCfg)

		result := FindConfigFile(labDir, "mylab")
		absParent, _ := filepath.Abs(parentCfg)
		if result != absParent {
			t.Errorf("expected parent config %q to take priority, got %q", absParent, result)
		}
	})

	t.Run("not found", func(t *testing.T) {
		result := FindConfigFile(filepath.Join(tmpDir, "nonexistent"), "mylab")
		if result != "" {
			t.Errorf("expected empty result, got %q", result)
		}
	})
}

func TestApplyConfig(t *testing.T) {
	// Build a base topology with label-based data
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{
				Name:  "spine-sw-01",
				Kind:  "linux",
				Image: "alpine:latest",
				Labels: map[string]string{
					"graph-dc":        "old-dc",
					"graph-rack":      "old-rack",
					"graph-rack-unit": "10",
				},
				Graph: GraphInfo{
					DC:           "old-dc",
					Rack:         "old-rack",
					RackUnit:     10,
					RackUnitSize: 1,
					Role:         "server",
					Icon:         "host",
				},
			},
			{
				Name:  "compute-01",
				Kind:  "linux",
				Image: "alpine:latest",
				Labels: map[string]string{},
				Graph: GraphInfo{
					RackUnitSize: 1,
					Role:         "server",
					Icon:         "server",
				},
			},
			{
				Name:  "untouched-node",
				Kind:  "linux",
				Image: "alpine:latest",
				Labels: map[string]string{
					"graph-dc":        "dc-a",
					"graph-rack":      "rack-a01",
					"graph-rack-unit": "20",
				},
				Graph: GraphInfo{
					DC:           "dc-a",
					Rack:         "rack-a01",
					RackUnit:     20,
					RackUnitSize: 1,
					Role:         "server",
					Icon:         "server",
				},
			},
		},
		Links: []Link{},
		Groups: Groups{
			DCs:   []string{"old-dc"},
			Racks: map[string][]string{"old-dc": {"old-rack"}},
		},
	}

	cfg := &Config{
		Racks: map[string]RackConfig{
			"rack-a01": {DC: "dc-a", Units: 42},
			"rack-a02": {DC: "dc-a", Units: 48},
		},
		Nodes: map[string]NodeConfig{
			"spine-sw-01": {Rack: "rack-a01", Unit: 42, Size: 1, Role: "spine"},
			"compute-01":  {Rack: "rack-a01", Unit: 39, Size: 2, Role: "server"},
		},
	}

	ApplyConfig(topo, cfg)

	// Check spine-sw-01 was overridden
	spine := topo.Nodes[0]
	if spine.Graph.DC != "dc-a" {
		t.Errorf("spine DC = %q, want %q", spine.Graph.DC, "dc-a")
	}
	if spine.Graph.Rack != "rack-a01" {
		t.Errorf("spine Rack = %q, want %q", spine.Graph.Rack, "rack-a01")
	}
	if spine.Graph.RackUnit != 42 {
		t.Errorf("spine RackUnit = %d, want %d", spine.Graph.RackUnit, 42)
	}
	if spine.Graph.Role != "spine" {
		t.Errorf("spine Role = %q, want %q", spine.Graph.Role, "spine")
	}

	// Check compute-01 was set
	compute := topo.Nodes[1]
	if compute.Graph.DC != "dc-a" {
		t.Errorf("compute DC = %q, want %q", compute.Graph.DC, "dc-a")
	}
	if compute.Graph.Rack != "rack-a01" {
		t.Errorf("compute Rack = %q, want %q", compute.Graph.Rack, "rack-a01")
	}
	if compute.Graph.RackUnit != 39 {
		t.Errorf("compute RackUnit = %d, want %d", compute.Graph.RackUnit, 39)
	}
	if compute.Graph.RackUnitSize != 2 {
		t.Errorf("compute RackUnitSize = %d, want %d", compute.Graph.RackUnitSize, 2)
	}

	// Check untouched-node was NOT modified
	untouched := topo.Nodes[2]
	if untouched.Graph.DC != "dc-a" {
		t.Errorf("untouched DC = %q, want %q", untouched.Graph.DC, "dc-a")
	}
	if untouched.Graph.RackUnit != 20 {
		t.Errorf("untouched RackUnit = %d, want %d", untouched.Graph.RackUnit, 20)
	}

	// Check rack units
	if topo.Groups.RackUnits["rack-a01"] != 42 {
		t.Errorf("rack-a01 units = %d, want 42", topo.Groups.RackUnits["rack-a01"])
	}
	if topo.Groups.RackUnits["rack-a02"] != 48 {
		t.Errorf("rack-a02 units = %d, want 48", topo.Groups.RackUnits["rack-a02"])
	}

	// Check groups were rebuilt
	if len(topo.Groups.DCs) != 1 || topo.Groups.DCs[0] != "dc-a" {
		t.Errorf("expected DCs = [dc-a], got %v", topo.Groups.DCs)
	}
	racks := topo.Groups.Racks["dc-a"]
	if len(racks) != 1 || racks[0] != "rack-a01" {
		t.Errorf("expected dc-a racks = [rack-a01], got %v", racks)
	}
}

func TestApplyConfigNil(t *testing.T) {
	topo := &Topology{
		Name:  "test",
		Nodes: []Node{},
		Groups: Groups{
			DCs:   []string{"dc1"},
			Racks: map[string][]string{"dc1": {"rack1"}},
		},
	}

	// Should not panic
	ApplyConfig(topo, nil)

	// Groups should be unchanged
	if len(topo.Groups.DCs) != 1 {
		t.Errorf("expected 1 DC, got %d", len(topo.Groups.DCs))
	}
}

func TestValidateLayoutClean(t *testing.T) {
	topo := &Topology{
		Nodes: []Node{
			{Name: "sw1", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 42, RackUnitSize: 1}},
			{Name: "sv1", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 10, RackUnitSize: 2}},
		},
		Groups: Groups{
			RackUnits: map[string]int{"rack1": 42},
		},
	}
	warns := ValidateLayout(topo)
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got %v", warns)
	}
}

func TestValidateLayoutExceedsRack(t *testing.T) {
	topo := &Topology{
		Nodes: []Node{
			{Name: "big", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 10, RackUnitSize: 4}},
		},
		Groups: Groups{
			RackUnits: map[string]int{"rack1": 12},
		},
	}
	warns := ValidateLayout(topo)
	// unit 10 + size 4 → U10–U13, rack is 12U → exceeds
	found := false
	for _, w := range warns {
		if contains(w, "exceeds") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'exceeds' warning, got %v", warns)
	}
}

func TestValidateLayoutOverlap(t *testing.T) {
	topo := &Topology{
		Nodes: []Node{
			{Name: "a", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 5, RackUnitSize: 2}},  // U5–U6
			{Name: "b", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 6, RackUnitSize: 1}},  // U6
		},
		Groups: Groups{
			RackUnits: map[string]int{"rack1": 42},
		},
	}
	warns := ValidateLayout(topo)
	found := false
	for _, w := range warns {
		if contains(w, "overlap") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'overlap' warning, got %v", warns)
	}
}

func TestValidateLayoutNoOverlapAdjacent(t *testing.T) {
	topo := &Topology{
		Nodes: []Node{
			{Name: "a", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 5, RackUnitSize: 2}},  // U5–U6
			{Name: "b", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 7, RackUnitSize: 1}},  // U7
		},
		Groups: Groups{
			RackUnits: map[string]int{"rack1": 42},
		},
	}
	warns := ValidateLayout(topo)
	for _, w := range warns {
		if contains(w, "overlap") {
			t.Errorf("expected no overlap warning, got %q", w)
		}
	}
}

func TestValidateLayoutMissingPlacement(t *testing.T) {
	topo := &Topology{
		Nodes: []Node{
			{Name: "ok", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 1, RackUnitSize: 1}},
			{Name: "nodc", Graph: GraphInfo{DC: "", Rack: "rack1", RackUnit: 2, RackUnitSize: 1}},
			{Name: "norack", Graph: GraphInfo{DC: "dc1", Rack: "", RackUnit: 3, RackUnitSize: 1}},
			{Name: "nounit", Graph: GraphInfo{DC: "dc1", Rack: "rack1", RackUnit: 0, RackUnitSize: 1}},
		},
		Groups: Groups{
			RackUnits: map[string]int{"rack1": 42},
		},
	}
	warns := ValidateLayout(topo)
	// Should warn for nodc, norack, nounit (3 warnings)
	count := 0
	for _, w := range warns {
		if contains(w, "missing rack placement") {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 missing placement warnings, got %d: %v", count, warns)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestRackUnitSizeLabel(t *testing.T) {
	data := []byte(`{
		"name": "size-test",
		"nodes": {
			"server1": {
				"shortname": "server1",
				"kind": "linux",
				"labels": {
					"graph-dc": "dc1",
					"graph-rack": "rack1",
					"graph-rack-unit": "10",
					"graph-rack-unit-size": "2"
				}
			},
			"switch1": {
				"shortname": "switch1",
				"kind": "nokia_srlinux",
				"labels": {
					"graph-dc": "dc1",
					"graph-rack": "rack1",
					"graph-rack-unit": "42"
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
		switch node.Name {
		case "server1":
			if node.Graph.RackUnitSize != 2 {
				t.Errorf("server1 RackUnitSize = %d, want 2", node.Graph.RackUnitSize)
			}
		case "switch1":
			if node.Graph.RackUnitSize != 1 {
				t.Errorf("switch1 RackUnitSize = %d, want 1 (default)", node.Graph.RackUnitSize)
			}
		}
	}
}
