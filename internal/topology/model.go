package topology

// Topology is the parsed representation of a clab topology
type Topology struct {
	Name   string `json:"name"`
	Nodes  []Node `json:"nodes"`
	Links  []Link `json:"links"`
	Groups Groups `json:"groups"`
}

// Node represents a containerlab node
type Node struct {
	Name          string            `json:"name"`
	ShortName     string            `json:"short_name"`
	LongName      string            `json:"long_name"`
	FQDN          string            `json:"fqdn"`
	Kind          string            `json:"kind"`
	Image         string            `json:"image"`
	Status        string            `json:"status"` // populated from Docker, not from topology-data.json
	MgmtIPv4      string            `json:"mgmt_ipv4"`
	MgmtIPv6      string            `json:"mgmt_ipv6"`
	ContainerID   string            `json:"container_id,omitempty"` // populated from Docker
	Labels        map[string]string `json:"labels,omitempty"`
	PortBindings  []PortBinding     `json:"port_bindings,omitempty"`
	AccessMethods []AccessMethod    `json:"access_methods,omitempty"`
	Graph         GraphInfo         `json:"graph"`
}

// PortBinding represents a port binding for a node
type PortBinding struct {
	HostIP   string `json:"host_ip"`
	HostPort int    `json:"host_port"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// AccessMethod represents an available access method for a node
type AccessMethod struct {
	Type   string `json:"type"`   // "exec", "ssh", "vnc"
	Label  string `json:"label"`
	Target string `json:"target"` // e.g., "172.20.20.2:22"
}

// Link represents a connection between two nodes
type Link struct {
	ID    string      `json:"id"`
	A     Endpoint    `json:"a"`
	Z     Endpoint    `json:"z"`
	State string      `json:"state"`        // "up", "down", "degraded" - populated from network state
	Netem *NetemConfig `json:"netem,omitempty"` // populated from network state
}

// Endpoint is one side of a link
type Endpoint struct {
	Node      string `json:"node"`
	Interface string `json:"interface"`
	MAC       string `json:"mac"`
}

// Groups represents the DC/rack grouping of nodes
type Groups struct {
	DCs   []string            `json:"dcs"`
	Racks map[string][]string `json:"racks"` // dc -> []rack
}

// GraphInfo holds graph layout information for a node
type GraphInfo struct {
	DC       string `json:"dc"`
	Rack     string `json:"rack"`
	RackUnit int    `json:"rack_unit"`
	Role     string `json:"role"`
	Icon     string `json:"icon"`
	Hidden   bool   `json:"hidden"`
}

// NetemConfig holds network emulation configuration
type NetemConfig struct {
	DelayMs          int     `json:"delay_ms"`
	JitterMs         int     `json:"jitter_ms"`
	LossPercent      float64 `json:"loss_percent"`
	CorruptPercent   float64 `json:"corrupt_percent"`
	DuplicatePercent float64 `json:"duplicate_percent"`
}
