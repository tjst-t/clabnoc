package topology

// SSHCredentials holds SSH authentication information.
type SSHCredentials struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Port     int    `json:"port" yaml:"port"`
}

// DefaultSSHCredentials maps containerlab kind names to their default SSH credentials.
var DefaultSSHCredentials = map[string]SSHCredentials{
	"nokia_srlinux": {Username: "admin", Password: "NokiaSrl1!", Port: 22},
	"srl":           {Username: "admin", Password: "NokiaSrl1!", Port: 22},
	"arista_ceos":   {Username: "admin", Password: "admin", Port: 22},
	"ceos":          {Username: "admin", Password: "admin", Port: 22},
	"sonic-vs":      {Username: "admin", Password: "YourPaSsWoRd", Port: 22},
	"linux":         {Username: "root", Password: "", Port: 22},
	"crpd":          {Username: "root", Password: "clab123", Port: 22},
	"vr-sros":       {Username: "admin", Password: "admin", Port: 22},
	"vr-vmx":        {Username: "root", Password: "Embe1mpls", Port: 22},
	"vr-xrv9k":      {Username: "clab", Password: "clab@123", Port: 22},
	"vr-veos":       {Username: "admin", Password: "admin", Port: 22},
	"vr-csr":        {Username: "admin", Password: "admin", Port: 22},
	"vr-n9kv":       {Username: "admin", Password: "admin", Port: 22},
}

// ResolveSSHCredentials resolves SSH credentials for a node using a 3-layer merge:
//  1. Built-in defaults for the kind
//  2. .clabnoc.yml kind_defaults override
//  3. .clabnoc.yml per-node override
//
// Each layer only overrides fields that are explicitly set (non-zero).
func ResolveSSHCredentials(kind, nodeName string, cfg *Config) SSHCredentials {
	// Layer 1: built-in defaults
	creds := SSHCredentials{
		Username: "admin",
		Password: "",
		Port:     22,
	}
	if builtin, ok := DefaultSSHCredentials[kind]; ok {
		creds = builtin
	}

	if cfg == nil {
		return creds
	}

	// Layer 2: kind_defaults from .clabnoc.yml
	if kc, ok := cfg.KindDefaults[kind]; ok && kc.SSH != nil {
		mergeSSHCredentials(&creds, kc.SSH)
	}

	// Layer 3: per-node override from .clabnoc.yml
	if nc, ok := cfg.Nodes[nodeName]; ok && nc.SSH != nil {
		mergeSSHCredentials(&creds, nc.SSH)
	}

	return creds
}

// mergeSSHCredentials applies non-zero fields from src onto dst.
func mergeSSHCredentials(dst *SSHCredentials, src *SSHCredentials) {
	if src.Username != "" {
		dst.Username = src.Username
	}
	if src.Password != "" {
		dst.Password = src.Password
	}
	if src.Port != 0 {
		dst.Port = src.Port
	}
}
