package topology

// RawTopology represents the raw topology-data.json structure from Containerlab.
type RawTopology struct {
	Name  string                 `json:"name"`
	Type  string                 `json:"type"`
	Clab  map[string]interface{} `json:"clab,omitempty"`
	Nodes map[string]RawNode     `json:"nodes"`
	Links []RawLink              `json:"links"`
}

// RawNode represents a node in topology-data.json.
type RawNode struct {
	Index              string            `json:"index"`
	ShortName          string            `json:"shortname"`
	LongName           string            `json:"longname"`
	FQDN               string            `json:"fqdn"`
	Group              string            `json:"group"`
	LabDir             string            `json:"labdir"`
	Kind               string            `json:"kind"`
	Image              string            `json:"image"`
	MgmtNet            string            `json:"mgmt-net"`
	MgmtIntf           string            `json:"mgmt-intf"`
	MgmtIPv4           string            `json:"mgmt-ipv4-address"`
	MgmtIPv4PrefixLen  int               `json:"mgmt-ipv4-prefix-length"`
	MgmtIPv6           string            `json:"mgmt-ipv6-address"`
	MgmtIPv6PrefixLen  int               `json:"mgmt-ipv6-prefix-length"`
	MACAddress         string            `json:"mac-address"`
	Labels             map[string]string  `json:"labels"`
	PortBindings       []PortBinding     `json:"port-bindings"`
}

// PortBinding represents a port binding entry.
type PortBinding struct {
	HostIP   string `json:"host-ip"`
	HostPort int    `json:"host-port"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// RawLink represents a link in topology-data.json.
// Supports both v0.73+ (endpoints wrapper) and legacy (flat a/z) formats.
type RawLink struct {
	// v0.73+ format
	Endpoints *RawEndpoints `json:"endpoints,omitempty"`
	// Legacy format
	A *RawEndpoint `json:"a,omitempty"`
	Z *RawEndpoint `json:"z,omitempty"`
}

// RawEndpoints wraps a/z endpoints (v0.73+ format).
type RawEndpoints struct {
	A RawEndpoint `json:"a"`
	Z RawEndpoint `json:"z"`
}

// RawEndpoint represents one side of a link.
type RawEndpoint struct {
	Node      string `json:"node"`
	Interface string `json:"interface"`
	MAC       string `json:"mac"`
	Peer      string `json:"peer,omitempty"`
}

// Topology is the processed topology ready for API responses.
type Topology struct {
	Name   string            `json:"name"`
	Nodes  []Node            `json:"nodes"`
	Links  []Link            `json:"links"`
	Groups Groups            `json:"groups"`
}

// Node represents a processed node for API responses.
type Node struct {
	Name          string            `json:"name"`
	Kind          string            `json:"kind"`
	Image         string            `json:"image"`
	Status        string            `json:"status"`
	MgmtIPv4     string            `json:"mgmt_ipv4"`
	MgmtIPv6     string            `json:"mgmt_ipv6"`
	ContainerID   string            `json:"container_id"`
	Labels        map[string]string `json:"labels"`
	PortBindings  []PortBinding     `json:"port_bindings"`
	AccessMethods []AccessMethod    `json:"access_methods"`
	Graph         GraphInfo         `json:"graph"`
}

// AccessMethod describes how to access a node.
type AccessMethod struct {
	Type   string `json:"type"`
	Label  string `json:"label"`
	Target string `json:"target,omitempty"`
}

// GraphInfo holds visualization metadata.
type GraphInfo struct {
	DC       string `json:"dc"`
	Rack     string `json:"rack"`
	RackUnit int    `json:"rack_unit"`
	Role     string `json:"role"`
	Icon     string `json:"icon"`
	Hidden   bool   `json:"hidden"`
}

// Link represents a processed link for API responses.
type Link struct {
	ID    string      `json:"id"`
	A     Endpoint    `json:"a"`
	Z     Endpoint    `json:"z"`
	State string      `json:"state"`
	Netem interface{} `json:"netem"`
}

// Endpoint represents one side of a link in API responses.
type Endpoint struct {
	Node      string `json:"node"`
	Interface string `json:"interface"`
	MAC       string `json:"mac,omitempty"`
}

// Groups holds the grouping structure for the topology.
type Groups struct {
	DCs   []string            `json:"dcs"`
	Racks map[string][]string `json:"racks"`
}
