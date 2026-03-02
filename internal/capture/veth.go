package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// VethResolver resolves container interface names to host-side veth peer names.
type VethResolver interface {
	Resolve(ctx context.Context, containerID, ifName string) (string, error)
}

// PIDProvider retrieves the PID of a container's init process.
type PIDProvider interface {
	GetPID(ctx context.Context, containerID string) (int, error)
}

// ProcVethResolver resolves veth peers by reading /proc and /sys filesystems.
type ProcVethResolver struct {
	pidProvider PIDProvider
	procRoot    string // defaults to "/proc"
	sysRoot     string // defaults to "/sys"
}

// NewProcVethResolver creates a new ProcVethResolver.
func NewProcVethResolver(pidProvider PIDProvider) *ProcVethResolver {
	return &ProcVethResolver{
		pidProvider: pidProvider,
		procRoot:    "/proc",
		sysRoot:     "/sys",
	}
}

// Resolve finds the host-side veth peer name for a container interface.
// Algorithm:
// 1. Get container PID via PIDProvider
// 2. Read /proc/{PID}/root/sys/class/net/{ifName}/iflink to get peer ifindex
// 3. Scan host /sys/class/net/*/ifindex to find matching veth name
func (r *ProcVethResolver) Resolve(ctx context.Context, containerID, ifName string) (string, error) {
	pid, err := r.pidProvider.GetPID(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("getting container PID: %w", err)
	}

	// Read the iflink from the container's network namespace via /proc
	iflinkPath := filepath.Join(r.procRoot, strconv.Itoa(pid), "root", "sys", "class", "net", ifName, "iflink")
	data, err := os.ReadFile(iflinkPath)
	if err != nil {
		return "", fmt.Errorf("reading iflink for %s: %w", ifName, err)
	}

	peerIndex, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return "", fmt.Errorf("parsing iflink value %q: %w", strings.TrimSpace(string(data)), err)
	}

	// Scan host /sys/class/net/*/ifindex to find the matching interface
	hostNetDir := filepath.Join(r.sysRoot, "class", "net")
	entries, err := os.ReadDir(hostNetDir)
	if err != nil {
		return "", fmt.Errorf("reading host net directory: %w", err)
	}

	for _, entry := range entries {
		ifindexPath := filepath.Join(hostNetDir, entry.Name(), "ifindex")
		idxData, err := os.ReadFile(ifindexPath)
		if err != nil {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(string(idxData)))
		if err != nil {
			continue
		}
		if idx == peerIndex {
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no host veth found with ifindex %d for container %s interface %s", peerIndex, containerID[:12], ifName)
}
