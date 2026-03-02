package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type mockPIDProvider struct {
	pid int
	err error
}

func (m *mockPIDProvider) GetPID(ctx context.Context, containerID string) (int, error) {
	return m.pid, m.err
}

func TestProcVethResolver(t *testing.T) {
	// Create a temp filesystem simulating /proc and /sys
	tmpDir := t.TempDir()
	procRoot := filepath.Join(tmpDir, "proc")
	sysRoot := filepath.Join(tmpDir, "sys")

	// Setup container's network namespace: /proc/1234/root/sys/class/net/eth1/iflink = "42"
	containerNetDir := filepath.Join(procRoot, "1234", "root", "sys", "class", "net", "eth1")
	if err := os.MkdirAll(containerNetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(containerNetDir, "iflink"), []byte("42\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Setup host interfaces:
	// /sys/class/net/lo/ifindex = "1"
	// /sys/class/net/vethABC123/ifindex = "42"
	// /sys/class/net/eth0/ifindex = "2"
	hostInterfaces := map[string]string{
		"lo":         "1",
		"vethABC123": "42",
		"eth0":       "2",
	}
	for ifName, ifIndex := range hostInterfaces {
		dir := filepath.Join(sysRoot, "class", "net", ifName)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "ifindex"), []byte(ifIndex+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	resolver := &ProcVethResolver{
		pidProvider: &mockPIDProvider{pid: 1234},
		procRoot:    procRoot,
		sysRoot:     sysRoot,
	}

	hostVeth, err := resolver.Resolve(context.Background(), "container-abc123def456", "eth1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostVeth != "vethABC123" {
		t.Errorf("expected vethABC123, got %s", hostVeth)
	}
}

func TestProcVethResolverPIDError(t *testing.T) {
	resolver := &ProcVethResolver{
		pidProvider: &mockPIDProvider{err: fmt.Errorf("container not running")},
		procRoot:    "/nonexistent",
		sysRoot:     "/nonexistent",
	}

	_, err := resolver.Resolve(context.Background(), "container-id", "eth0")
	if err == nil {
		t.Error("expected error")
	}
}

func TestProcVethResolverIflinkNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	procRoot := filepath.Join(tmpDir, "proc")

	// Create proc dir but no iflink file
	if err := os.MkdirAll(filepath.Join(procRoot, "1234", "root", "sys", "class", "net"), 0o755); err != nil {
		t.Fatal(err)
	}

	resolver := &ProcVethResolver{
		pidProvider: &mockPIDProvider{pid: 1234},
		procRoot:    procRoot,
		sysRoot:     filepath.Join(tmpDir, "sys"),
	}

	_, err := resolver.Resolve(context.Background(), "container-id", "eth0")
	if err == nil {
		t.Error("expected error for missing iflink")
	}
}

func TestProcVethResolverNoMatchingHost(t *testing.T) {
	tmpDir := t.TempDir()
	procRoot := filepath.Join(tmpDir, "proc")
	sysRoot := filepath.Join(tmpDir, "sys")

	// Container iflink = 99 (no host interface has this)
	containerNetDir := filepath.Join(procRoot, "1234", "root", "sys", "class", "net", "eth1")
	if err := os.MkdirAll(containerNetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(containerNetDir, "iflink"), []byte("99\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Host has only lo
	loDir := filepath.Join(sysRoot, "class", "net", "lo")
	if err := os.MkdirAll(loDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loDir, "ifindex"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	resolver := &ProcVethResolver{
		pidProvider: &mockPIDProvider{pid: 1234},
		procRoot:    procRoot,
		sysRoot:     sysRoot,
	}

	_, err := resolver.Resolve(context.Background(), "container-abc123def456", "eth1")
	if err == nil {
		t.Error("expected error for no matching host veth")
	}
}
