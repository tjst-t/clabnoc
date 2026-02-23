package topology

import (
	"encoding/json"
	"fmt"
)

// topoDataRaw is the raw topology-data.json structure
type topoDataRaw struct {
	Name  string                 `json:"name"`
	Type  string                 `json:"type"`
	Clab  *topoClabRaw           `json:"clab"`
	Nodes map[string]nodeRaw     `json:"nodes"`
	Links []json.RawMessage      `json:"links"` // handle both formats
}

type topoClabRaw struct {
	Config *topoConfigRaw `json:"config"`
}

type topoConfigRaw struct {
	Prefix string       `json:"prefix"`
	Mgmt   *topoMgmtRaw `json:"mgmt"`
}

type topoMgmtRaw struct {
	Network string `json:"network"`
	Bridge  string `json:"bridge"`
}

type nodeRaw struct {
	Index                string            `json:"index"`
	ShortName            string            `json:"shortname"`
	LongName             string            `json:"longname"`
	FQDN                 string            `json:"fqdn"`
	Group                string            `json:"group"`
	LabDir               string            `json:"labdir"`
	Kind                 string            `json:"kind"`
	Image                string            `json:"image"`
	MgmtNet              string            `json:"mgmt-net"`
	MgmtIntf             string            `json:"mgmt-intf"`
	MgmtIPv4Address      string            `json:"mgmt-ipv4-address"`
	MgmtIPv4PrefixLength int               `json:"mgmt-ipv4-prefix-length"`
	MgmtIPv6Address      string            `json:"mgmt-ipv6-address"`
	MgmtIPv6PrefixLength int               `json:"mgmt-ipv6-prefix-length"`
	MACAddress           string            `json:"mac-address"`
	Labels               map[string]string `json:"labels"`
	PortBindings         []portBindingRaw  `json:"port-bindings"`
}

type portBindingRaw struct {
	HostIP   string `json:"host-ip"`
	HostPort int    `json:"host-port"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// linkNewFormat v0.73.0+
type linkNewFormat struct {
	Endpoints struct {
		A endpointRaw `json:"a"`
		Z endpointRaw `json:"z"`
	} `json:"endpoints"`
}

// linkLegacyFormat legacy format
type linkLegacyFormat struct {
	A endpointRaw `json:"a"`
	Z endpointRaw `json:"z"`
}

type endpointRaw struct {
	Node      string `json:"node"`
	Interface string `json:"interface"`
	MAC       string `json:"mac"`
	Peer      string `json:"peer,omitempty"`
}

// parseLink parses a raw JSON link, handling both formats
func parseLink(raw json.RawMessage) (Link, error) {
	// Try new format first (has "endpoints" key)
	var newFmt linkNewFormat
	if err := json.Unmarshal(raw, &newFmt); err == nil && newFmt.Endpoints.A.Node != "" {
		return Link{
			A: Endpoint{
				Node:      newFmt.Endpoints.A.Node,
				Interface: newFmt.Endpoints.A.Interface,
				MAC:       newFmt.Endpoints.A.MAC,
			},
			Z: Endpoint{
				Node:      newFmt.Endpoints.Z.Node,
				Interface: newFmt.Endpoints.Z.Interface,
				MAC:       newFmt.Endpoints.Z.MAC,
			},
		}, nil
	}

	// Fall back to legacy format
	var legFmt linkLegacyFormat
	if err := json.Unmarshal(raw, &legFmt); err != nil {
		return Link{}, fmt.Errorf("unmarshal legacy link format: %w", err)
	}
	if legFmt.A.Node == "" {
		return Link{}, fmt.Errorf("link has no node information")
	}
	return Link{
		A: Endpoint{
			Node:      legFmt.A.Node,
			Interface: legFmt.A.Interface,
			MAC:       legFmt.A.MAC,
		},
		Z: Endpoint{
			Node:      legFmt.Z.Node,
			Interface: legFmt.Z.Interface,
			MAC:       legFmt.Z.MAC,
		},
	}, nil
}
