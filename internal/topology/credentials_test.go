package topology

import "testing"

func TestResolveSSHCredentials(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		nodeName string
		cfg      *Config
		want     SSHCredentials
	}{
		{
			name:     "builtin default for nokia_srlinux",
			kind:     "nokia_srlinux",
			nodeName: "srl1",
			cfg:      nil,
			want:     SSHCredentials{Username: "admin", Password: "NokiaSrl1!", Port: 22},
		},
		{
			name:     "builtin default for srl alias",
			kind:     "srl",
			nodeName: "srl1",
			cfg:      nil,
			want:     SSHCredentials{Username: "admin", Password: "NokiaSrl1!", Port: 22},
		},
		{
			name:     "builtin default for arista_ceos",
			kind:     "arista_ceos",
			nodeName: "ceos1",
			cfg:      nil,
			want:     SSHCredentials{Username: "admin", Password: "admin", Port: 22},
		},
		{
			name:     "builtin default for linux",
			kind:     "linux",
			nodeName: "host1",
			cfg:      nil,
			want:     SSHCredentials{Username: "root", Password: "", Port: 22},
		},
		{
			name:     "unknown kind uses generic default",
			kind:     "unknown_kind",
			nodeName: "node1",
			cfg:      nil,
			want:     SSHCredentials{Username: "admin", Password: "", Port: 22},
		},
		{
			name:     "nil config returns builtin",
			kind:     "nokia_srlinux",
			nodeName: "srl1",
			cfg:      nil,
			want:     SSHCredentials{Username: "admin", Password: "NokiaSrl1!", Port: 22},
		},
		{
			name:     "kind_defaults override",
			kind:     "nokia_srlinux",
			nodeName: "srl1",
			cfg: &Config{
				KindDefaults: map[string]KindConfig{
					"nokia_srlinux": {
						SSH: &SSHCredentials{Username: "operator", Password: "secret123", Port: 2222},
					},
				},
				Nodes: map[string]NodeConfig{},
			},
			want: SSHCredentials{Username: "operator", Password: "secret123", Port: 2222},
		},
		{
			name:     "node override takes precedence over kind_defaults",
			kind:     "nokia_srlinux",
			nodeName: "srl1",
			cfg: &Config{
				KindDefaults: map[string]KindConfig{
					"nokia_srlinux": {
						SSH: &SSHCredentials{Username: "operator", Password: "secret123"},
					},
				},
				Nodes: map[string]NodeConfig{
					"srl1": {
						SSH: &SSHCredentials{Username: "nodeuser", Password: "nodepass"},
					},
				},
			},
			want: SSHCredentials{Username: "nodeuser", Password: "nodepass", Port: 22},
		},
		{
			name:     "partial kind_defaults override (only password)",
			kind:     "nokia_srlinux",
			nodeName: "srl1",
			cfg: &Config{
				KindDefaults: map[string]KindConfig{
					"nokia_srlinux": {
						SSH: &SSHCredentials{Password: "newpass"},
					},
				},
				Nodes: map[string]NodeConfig{},
			},
			want: SSHCredentials{Username: "admin", Password: "newpass", Port: 22},
		},
		{
			name:     "partial node override (only port)",
			kind:     "nokia_srlinux",
			nodeName: "srl1",
			cfg: &Config{
				KindDefaults: map[string]KindConfig{},
				Nodes: map[string]NodeConfig{
					"srl1": {
						SSH: &SSHCredentials{Port: 2222},
					},
				},
			},
			want: SSHCredentials{Username: "admin", Password: "NokiaSrl1!", Port: 2222},
		},
		{
			name:     "empty config returns builtin",
			kind:     "arista_ceos",
			nodeName: "ceos1",
			cfg: &Config{
				KindDefaults: map[string]KindConfig{},
				Nodes:        map[string]NodeConfig{},
			},
			want: SSHCredentials{Username: "admin", Password: "admin", Port: 22},
		},
		{
			name:     "node override without kind_defaults",
			kind:     "linux",
			nodeName: "host1",
			cfg: &Config{
				Nodes: map[string]NodeConfig{
					"host1": {
						SSH: &SSHCredentials{Username: "ubuntu", Password: "ubuntu"},
					},
				},
			},
			want: SSHCredentials{Username: "ubuntu", Password: "ubuntu", Port: 22},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveSSHCredentials(tt.kind, tt.nodeName, tt.cfg)
			if got.Username != tt.want.Username {
				t.Errorf("Username = %q, want %q", got.Username, tt.want.Username)
			}
			if got.Password != tt.want.Password {
				t.Errorf("Password = %q, want %q", got.Password, tt.want.Password)
			}
			if got.Port != tt.want.Port {
				t.Errorf("Port = %d, want %d", got.Port, tt.want.Port)
			}
		})
	}
}
