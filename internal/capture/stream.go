package capture

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// PacketInfo represents a parsed packet from tcpdump output.
type PacketInfo struct {
	No          int    `json:"no"`
	Time        string `json:"time"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Protocol    string `json:"protocol"`
	Length      int    `json:"length"`
	Info        string `json:"info"`
}

// tcpdump -nn output patterns:
// TCP: 12:34:56.789012 IP 10.0.0.1.443 > 10.0.0.2.52341: Flags [S.], seq 1234, ack 5678, win 65535, length 0
// UDP: 12:34:56.789012 IP 10.0.0.1.53 > 10.0.0.2.12345: UDP, length 64
// ICMP: 12:34:56.789012 IP 10.0.0.1 > 10.0.0.2: ICMP echo request, id 1234, seq 1, length 64
// ARP: 12:34:56.789012 ARP, Request who-has 10.0.0.1 tell 10.0.0.2, length 28

var (
	// Matches IP packets: timestamp IP src > dst: protocol info
	ipPacketRe = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}\.\d+)\s+IP6?\s+(\S+)\s+>\s+(\S+?):\s+(.*)$`)

	// Matches ARP packets: timestamp ARP, ...
	arpPacketRe = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}\.\d+)\s+ARP,\s+(.*)$`)

	// Extract "length N" from packet info
	lengthRe = regexp.MustCompile(`length\s+(\d+)`)
)

// ParseTcpdumpLine parses a single line of tcpdump -nn output into a PacketInfo.
func ParseTcpdumpLine(line string, seqNo int) (*PacketInfo, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	// Try ARP first (simpler pattern)
	if m := arpPacketRe.FindStringSubmatch(line); m != nil {
		info := m[2]
		length := extractLength(info)

		// Parse ARP: "Request who-has X tell Y" or "Reply X is-at Y"
		src, dst := parseARPEndpoints(info)

		return &PacketInfo{
			No:          seqNo,
			Time:        m[1],
			Source:      src,
			Destination: dst,
			Protocol:    "ARP",
			Length:      length,
			Info:        info,
		}, nil
	}

	// Try IP packet
	if m := ipPacketRe.FindStringSubmatch(line); m != nil {
		timestamp := m[1]
		src := m[2]
		dst := m[3]
		rest := m[4]

		proto, info := classifyIPPacket(rest)
		length := extractLength(rest)

		// Clean up addresses: remove trailing port dot notation for display
		src = cleanAddress(src)
		dst = cleanAddress(dst)

		return &PacketInfo{
			No:          seqNo,
			Time:        timestamp,
			Source:      src,
			Destination: dst,
			Protocol:    proto,
			Length:      length,
			Info:        info,
		}, nil
	}

	return nil, fmt.Errorf("unrecognized tcpdump line: %s", line)
}

func classifyIPPacket(rest string) (protocol, info string) {
	rest = strings.TrimSpace(rest)

	switch {
	case strings.HasPrefix(rest, "Flags ["):
		// TCP packet
		return "TCP", rest
	case strings.HasPrefix(rest, "UDP,"):
		return "UDP", rest
	case strings.HasPrefix(rest, "ICMP"):
		return "ICMP", rest
	case strings.HasPrefix(rest, "ICMP6"):
		return "ICMPv6", rest
	default:
		// Try to identify by content
		if strings.Contains(rest, "UDP") {
			return "UDP", rest
		}
		if strings.Contains(rest, "ICMP") {
			return "ICMP", rest
		}
		// Unknown protocol, show the raw info
		return "IP", rest
	}
}

func extractLength(s string) int {
	if m := lengthRe.FindStringSubmatch(s); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}
	return 0
}

func cleanAddress(addr string) string {
	// Remove trailing colon if present
	addr = strings.TrimSuffix(addr, ":")
	return addr
}

var arpRequestRe = regexp.MustCompile(`Request who-has ([^\s,]+)\s+tell\s+([^\s,]+)`)
var arpReplyRe = regexp.MustCompile(`Reply ([^\s,]+) is-at ([^\s,]+)`)

func parseARPEndpoints(info string) (src, dst string) {
	if m := arpRequestRe.FindStringSubmatch(info); m != nil {
		return m[2], m[1] // tell X -> who-has Y: src=teller, dst=target
	}
	if m := arpReplyRe.FindStringSubmatch(info); m != nil {
		return m[1], m[2] // Reply X is-at Y: src=IP, dst=MAC
	}
	return "", ""
}

// StreamExecutor abstracts the execution of tcpdump for live streaming.
type StreamExecutor interface {
	StartStream(ctx context.Context, iface, bpfFilter string) (io.ReadCloser, *exec.Cmd, error)
}

// HostStreamExecutor runs tcpdump on the host for streaming.
type HostStreamExecutor struct{}

// StartStream begins a line-buffered tcpdump capture for streaming.
func (e *HostStreamExecutor) StartStream(ctx context.Context, iface, bpfFilter string) (io.ReadCloser, *exec.Cmd, error) {
	args := []string{"-l", "-nn", "-i", iface}
	if bpfFilter != "" {
		args = append(args, bpfFilter)
	}

	cmd := exec.CommandContext(ctx, "tcpdump", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("starting tcpdump stream: %w", err)
	}

	return stdout, cmd, nil
}
