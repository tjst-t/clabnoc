package capture

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type mockCaptureExecutor struct {
	startErr error
	stopErr  error
	// Track calls
	startCalls []startCall
	stopCalls  int
}

type startCall struct {
	Iface     string
	FilePath  string
	BPFFilter string
}

func (m *mockCaptureExecutor) Start(ctx context.Context, iface, filePath, bpfFilter string) (*exec.Cmd, error) {
	m.startCalls = append(m.startCalls, startCall{iface, filePath, bpfFilter})
	if m.startErr != nil {
		return nil, m.startErr
	}
	// Start a real (but harmless) process so we can wait on it
	cmd := exec.CommandContext(ctx, "sleep", "60")
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func (m *mockCaptureExecutor) Stop(cmd *exec.Cmd) error {
	m.stopCalls++
	if m.stopErr != nil {
		return m.stopErr
	}
	if cmd.Process != nil {
		return cmd.Process.Signal(os.Interrupt)
	}
	return nil
}

func TestCaptureManagerStart(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{}
	mgr := NewCaptureManager(executor, tmpDir)

	session, err := mgr.Start(context.Background(), "link1", "veth123", "tcp port 80")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.LinkID != "link1" {
		t.Errorf("expected link ID link1, got %s", session.LinkID)
	}
	if session.Interface != "veth123" {
		t.Errorf("expected interface veth123, got %s", session.Interface)
	}
	if session.BPFFilter != "tcp port 80" {
		t.Errorf("expected BPF filter 'tcp port 80', got %s", session.BPFFilter)
	}
	if !session.Active {
		t.Error("expected session to be active")
	}
	if session.FilePath == "" {
		t.Error("expected non-empty file path")
	}

	if len(executor.startCalls) != 1 {
		t.Fatalf("expected 1 start call, got %d", len(executor.startCalls))
	}
	if executor.startCalls[0].Iface != "veth123" {
		t.Errorf("expected iface veth123, got %s", executor.startCalls[0].Iface)
	}
	if executor.startCalls[0].BPFFilter != "tcp port 80" {
		t.Errorf("expected BPF filter, got %s", executor.startCalls[0].BPFFilter)
	}

	// Cleanup
	_ = mgr.Stop("link1")
}

func TestCaptureManagerDoubleStart(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{}
	mgr := NewCaptureManager(executor, tmpDir)

	_, err := mgr.Start(context.Background(), "link1", "veth123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = mgr.Start(context.Background(), "link1", "veth123", "")
	if err == nil {
		t.Error("expected error on double start")
	}

	_ = mgr.Stop("link1")
}

func TestCaptureManagerStop(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{}
	mgr := NewCaptureManager(executor, tmpDir)

	_, err := mgr.Start(context.Background(), "link1", "veth123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = mgr.Stop("link1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if executor.stopCalls != 1 {
		t.Errorf("expected 1 stop call, got %d", executor.stopCalls)
	}

	// Wait a bit for goroutine to mark session inactive
	time.Sleep(100 * time.Millisecond)

	session := mgr.GetSession("link1")
	if session == nil {
		t.Fatal("expected session to exist after stop")
	}
	if session.Active {
		t.Error("expected session to be inactive after stop")
	}
}

func TestCaptureManagerStopNotActive(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{}
	mgr := NewCaptureManager(executor, tmpDir)

	err := mgr.Stop("link1")
	if err == nil {
		t.Error("expected error when stopping non-existent capture")
	}
}

func TestCaptureManagerGetSession(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{}
	mgr := NewCaptureManager(executor, tmpDir)

	session := mgr.GetSession("link1")
	if session != nil {
		t.Error("expected nil session for non-existent link")
	}

	_, _ = mgr.Start(context.Background(), "link1", "veth123", "")
	session = mgr.GetSession("link1")
	if session == nil {
		t.Error("expected non-nil session after start")
	}

	_ = mgr.Stop("link1")
}

func TestCaptureManagerGetFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{}
	mgr := NewCaptureManager(executor, tmpDir)

	_, err := mgr.GetFilePath("link1")
	if err == nil {
		t.Error("expected error for non-existent session")
	}

	_, _ = mgr.Start(context.Background(), "link1", "veth123", "")
	filePath, err := mgr.GetFilePath("link1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filePath == "" {
		t.Error("expected non-empty file path")
	}
	if filepath.Dir(filePath) != tmpDir {
		t.Errorf("expected file in %s, got %s", tmpDir, filepath.Dir(filePath))
	}

	_ = mgr.Stop("link1")
}

func TestCaptureManagerCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{}
	mgr := NewCaptureManager(executor, tmpDir)

	_, _ = mgr.Start(context.Background(), "link1", "veth123", "")

	// Can't cleanup active capture
	err := mgr.Cleanup("link1")
	if err == nil {
		t.Error("expected error cleaning up active capture")
	}

	_ = mgr.Stop("link1")
	time.Sleep(100 * time.Millisecond)

	err = mgr.Cleanup("link1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	session := mgr.GetSession("link1")
	if session != nil {
		t.Error("expected session to be removed after cleanup")
	}
}

func TestCaptureManagerStartError(t *testing.T) {
	tmpDir := t.TempDir()
	executor := &mockCaptureExecutor{startErr: fmt.Errorf("tcpdump not found")}
	mgr := NewCaptureManager(executor, tmpDir)

	_, err := mgr.Start(context.Background(), "link1", "veth123", "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"link1", "link1"},
		{"spine1:e1-1__leaf1:e1-49", "spine1_e1-1__leaf1_e1-49"},
		{"a/b\\c.d", "a_b_c_d"},
	}
	for _, tt := range tests {
		got := sanitizeFileName(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeFileName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
