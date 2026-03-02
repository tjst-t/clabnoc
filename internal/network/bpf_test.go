package network

import (
	"testing"
)

func TestBPFPresets(t *testing.T) {
	presets := BPFPresets()

	if len(presets) == 0 {
		t.Fatal("expected at least one preset")
	}

	expectedNames := []string{"DNS", "BGP", "HTTPS", "HTTP", "ICMP", "ARP"}
	if len(presets) != len(expectedNames) {
		t.Fatalf("expected %d presets, got %d", len(expectedNames), len(presets))
	}

	for i, name := range expectedNames {
		if presets[i].Name != name {
			t.Errorf("preset[%d]: expected name %q, got %q", i, name, presets[i].Name)
		}
		if presets[i].Filter == "" {
			t.Errorf("preset[%d] %q: filter is empty", i, name)
		}
		if presets[i].Description == "" {
			t.Errorf("preset[%d] %q: description is empty", i, name)
		}
	}
}

func TestBuildTCFilterRules(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantCount int
		wantProto string
		wantErr   bool
	}{
		{
			name:      "tcp port",
			expr:      "tcp port 80",
			wantCount: 2, // src + dst rules
			wantProto: "ip",
		},
		{
			name:      "udp port",
			expr:      "udp port 53",
			wantCount: 2,
			wantProto: "ip",
		},
		{
			name:      "icmp",
			expr:      "icmp",
			wantCount: 1,
			wantProto: "ip",
		},
		{
			name:      "arp",
			expr:      "arp",
			wantCount: 1,
			wantProto: "0x0806",
		},
		{
			name:      "tcp only (no port)",
			expr:      "tcp",
			wantCount: 1,
			wantProto: "ip",
		},
		{
			name:      "udp only (no port)",
			expr:      "udp",
			wantCount: 1,
			wantProto: "ip",
		},
		{
			name:    "empty expression",
			expr:    "",
			wantErr: true,
		},
		{
			name:    "unsupported expression",
			expr:    "vlan 100",
			wantErr: true,
		},
		{
			name:    "invalid port",
			expr:    "tcp port abc",
			wantErr: true,
		},
		{
			name:    "port out of range",
			expr:    "tcp port 99999",
			wantErr: true,
		},
		{
			name:      "or combination",
			expr:      "tcp port 80 or tcp port 443",
			wantCount: 4, // 2 rules per port
			wantProto: "ip",
		},
		{
			name:      "mixed or",
			expr:      "icmp or arp",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := BuildTCFilterRules(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q, got nil", tt.expr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(rules) != tt.wantCount {
				t.Errorf("expected %d rules, got %d", tt.wantCount, len(rules))
			}
			if tt.wantProto != "" && len(rules) > 0 {
				if rules[0].Protocol != tt.wantProto {
					t.Errorf("expected protocol %q, got %q", tt.wantProto, rules[0].Protocol)
				}
			}
		})
	}
}

func TestBuildTCFilterCommands(t *testing.T) {
	cmds, err := BuildTCFilterCommands("eth0", "tcp port 179")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands (src+dst), got %d", len(cmds))
	}

	// Verify command structure
	for i, cmd := range cmds {
		// Should start with "filter add dev eth0 parent 1:0"
		if cmd[0] != "filter" || cmd[1] != "add" || cmd[2] != "dev" || cmd[3] != "eth0" {
			t.Errorf("cmd[%d]: unexpected prefix: %v", i, cmd[:4])
		}
		if cmd[4] != "parent" || cmd[5] != "1:0" {
			t.Errorf("cmd[%d]: expected parent 1:0, got %s %s", i, cmd[4], cmd[5])
		}
		// Should end with "flowid 1:1"
		if cmd[len(cmd)-2] != "flowid" || cmd[len(cmd)-1] != "1:1" {
			t.Errorf("cmd[%d]: expected flowid 1:1 at end, got %s %s", i, cmd[len(cmd)-2], cmd[len(cmd)-1])
		}
	}
}

func TestBuildTCFilterCommandsICMP(t *testing.T) {
	cmds, err := BuildTCFilterCommands("veth123", "icmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}

	cmd := cmds[0]
	if cmd[3] != "veth123" {
		t.Errorf("expected interface veth123, got %s", cmd[3])
	}
}

func TestBuildTCFilterCommandsError(t *testing.T) {
	_, err := BuildTCFilterCommands("eth0", "")
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestPresetFiltersAreValid(t *testing.T) {
	for _, preset := range BPFPresets() {
		t.Run(preset.Name, func(t *testing.T) {
			rules, err := BuildTCFilterRules(preset.Filter)
			if err != nil {
				t.Errorf("preset %q filter %q is invalid: %v", preset.Name, preset.Filter, err)
			}
			if len(rules) == 0 {
				t.Errorf("preset %q filter %q produced no rules", preset.Name, preset.Filter)
			}
		})
	}
}
