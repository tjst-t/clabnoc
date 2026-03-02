package network

import (
	"fmt"
	"strconv"
	"strings"
)

// BPFPreset represents a predefined BPF filter expression for common protocols.
type BPFPreset struct {
	Name        string `json:"name"`
	Filter      string `json:"filter"`
	Description string `json:"description"`
}

// BPFPresets returns the list of built-in BPF filter presets.
func BPFPresets() []BPFPreset {
	return []BPFPreset{
		{Name: "DNS", Filter: "udp port 53", Description: "DNS queries and responses (UDP/53)"},
		{Name: "BGP", Filter: "tcp port 179", Description: "BGP sessions (TCP/179)"},
		{Name: "HTTPS", Filter: "tcp port 443", Description: "HTTPS traffic (TCP/443)"},
		{Name: "HTTP", Filter: "tcp port 80", Description: "HTTP traffic (TCP/80)"},
		{Name: "ICMP", Filter: "icmp", Description: "ICMP echo, unreachable, etc."},
		{Name: "ARP", Filter: "arp", Description: "ARP requests and replies"},
	}
}

// TCFilterRule holds the arguments for a single tc filter add command.
type TCFilterRule struct {
	// Args is the list of arguments after "tc filter add dev <iface> parent 1:0 protocol <proto>"
	Protocol string
	Matches  []string
}

// BuildTCFilterRules converts a simple BPF-style filter expression into tc filter u32 match rules.
// Supports: "tcp port N", "udp port N", "icmp", "arp", and combinations with "or".
// For unsupported expressions, returns an error suggesting the use of tcpdump -ddd.
func BuildTCFilterRules(bpfExpr string) ([]TCFilterRule, error) {
	bpfExpr = strings.TrimSpace(bpfExpr)
	if bpfExpr == "" {
		return nil, fmt.Errorf("empty BPF filter expression")
	}

	// Split by " or " for disjunctions
	parts := splitOr(bpfExpr)
	var rules []TCFilterRule
	for _, part := range parts {
		r, err := parseSingleFilter(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("parsing filter %q: %w", part, err)
		}
		rules = append(rules, r...)
	}
	return rules, nil
}

func splitOr(expr string) []string {
	// Split on " or " (case-insensitive word boundary)
	var parts []string
	lower := strings.ToLower(expr)
	for {
		idx := strings.Index(lower, " or ")
		if idx < 0 {
			parts = append(parts, expr)
			break
		}
		parts = append(parts, expr[:idx])
		expr = expr[idx+4:]
		lower = lower[idx+4:]
	}
	return parts
}

func parseSingleFilter(expr string) ([]TCFilterRule, error) {
	tokens := strings.Fields(strings.ToLower(expr))
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty filter")
	}

	switch tokens[0] {
	case "icmp":
		// IP protocol 1 (ICMP)
		return []TCFilterRule{{
			Protocol: "ip",
			Matches:  []string{"u32", "match", "ip", "protocol", "1", "0xff"},
		}}, nil

	case "arp":
		// ARP is ethertype 0x0806
		return []TCFilterRule{{
			Protocol: "0x0806",
			Matches:  []string{"u32", "match", "u32", "0", "0"},
		}}, nil

	case "tcp", "udp":
		return parseTransportFilter(tokens)

	default:
		return nil, fmt.Errorf("unsupported filter expression %q: use tcpdump-style filters like 'tcp port 80', 'udp port 53', 'icmp', or 'arp'", expr)
	}
}

func parseTransportFilter(tokens []string) ([]TCFilterRule, error) {
	proto := tokens[0]
	var protoNum string
	if proto == "tcp" {
		protoNum = "6"
	} else {
		protoNum = "17"
	}

	if len(tokens) < 3 {
		// Just "tcp" or "udp" without port
		return []TCFilterRule{{
			Protocol: "ip",
			Matches:  []string{"u32", "match", "ip", "protocol", protoNum, "0xff"},
		}}, nil
	}

	if tokens[1] != "port" {
		return nil, fmt.Errorf("expected 'port' after %q, got %q", proto, tokens[1])
	}

	port, err := strconv.Atoi(tokens[2])
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port number: %q", tokens[2])
	}

	// For "tcp/udp port N", we need two rules: one for src port, one for dst port
	portHex := fmt.Sprintf("0x%04x", port)

	return []TCFilterRule{
		{
			Protocol: "ip",
			Matches: []string{"u32",
				"match", "ip", "protocol", protoNum, "0xff",
				"match", "ip", "sport", portHex, "0xffff",
			},
		},
		{
			Protocol: "ip",
			Matches: []string{"u32",
				"match", "ip", "protocol", protoNum, "0xff",
				"match", "ip", "dport", portHex, "0xffff",
			},
		},
	}, nil
}

// BuildTCFilterCommands generates the full tc filter add command arguments for a given interface and BPF expression.
// Each returned slice is the arguments for a single tc filter command (without the "tc" prefix).
func BuildTCFilterCommands(ifName, bpfExpr string) ([][]string, error) {
	rules, err := BuildTCFilterRules(bpfExpr)
	if err != nil {
		return nil, err
	}

	var cmds [][]string
	for _, r := range rules {
		cmd := []string{
			"filter", "add", "dev", ifName,
			"parent", "1:0",
			"protocol", r.Protocol,
			"prio", "1",
		}
		cmd = append(cmd, r.Matches...)
		cmd = append(cmd, "flowid", "1:1")
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}
