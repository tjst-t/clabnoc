package capture

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// CaptureSession represents an active or completed packet capture session.
type CaptureSession struct {
	ID        string    `json:"id"`
	LinkID    string    `json:"link_id"`
	Interface string    `json:"interface"`
	StartTime time.Time `json:"start_time"`
	FilePath  string    `json:"file_path"`
	BPFFilter string    `json:"bpf_filter,omitempty"`
	Active    bool      `json:"active"`

	cmd    *exec.Cmd
	cancel context.CancelFunc
	done   chan struct{}
}

// CaptureExecutor abstracts the execution of tcpdump for capture.
type CaptureExecutor interface {
	Start(ctx context.Context, iface, filePath, bpfFilter string) (*exec.Cmd, error)
	Stop(cmd *exec.Cmd) error
}

// HostCaptureExecutor runs tcpdump on the host.
type HostCaptureExecutor struct{}

// Start begins a tcpdump capture writing to a pcap file.
func (e *HostCaptureExecutor) Start(ctx context.Context, iface, filePath, bpfFilter string) (*exec.Cmd, error) {
	args := []string{"-i", iface, "-w", filePath, "-n"}
	if bpfFilter != "" {
		args = append(args, bpfFilter)
	}

	cmd := exec.CommandContext(ctx, "tcpdump", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting tcpdump: %w", err)
	}
	return cmd, nil
}

// Stop sends SIGINT to tcpdump for graceful shutdown.
func (e *HostCaptureExecutor) Stop(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(os.Interrupt)
}

// CaptureManager manages packet capture sessions.
type CaptureManager struct {
	executor CaptureExecutor
	sessions map[string]*CaptureSession // keyed by linkID
	mu       sync.Mutex
	baseDir  string
}

// NewCaptureManager creates a new CaptureManager.
func NewCaptureManager(executor CaptureExecutor, baseDir string) *CaptureManager {
	return &CaptureManager{
		executor: executor,
		sessions: make(map[string]*CaptureSession),
		baseDir:  baseDir,
	}
}

// Start begins a packet capture on the given link interface.
func (m *CaptureManager) Start(ctx context.Context, linkID, iface, bpfFilter string) (*CaptureSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[linkID]; ok && session.Active {
		return nil, fmt.Errorf("capture already active on link %s", linkID)
	}

	if err := os.MkdirAll(m.baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating capture directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("capture-%s-%s.pcap", sanitizeFileName(linkID), timestamp)
	filePath := filepath.Join(m.baseDir, fileName)

	captureCtx, cancel := context.WithCancel(ctx)
	cmd, err := m.executor.Start(captureCtx, iface, filePath, bpfFilter)
	if err != nil {
		cancel()
		return nil, err
	}

	session := &CaptureSession{
		ID:        fmt.Sprintf("%s-%s", linkID, timestamp),
		LinkID:    linkID,
		Interface: iface,
		StartTime: time.Now(),
		FilePath:  filePath,
		BPFFilter: bpfFilter,
		Active:    true,
		cmd:       cmd,
		cancel:    cancel,
		done:      make(chan struct{}),
	}

	// Wait for process to finish in background
	go func() {
		defer close(session.done)
		if err := cmd.Wait(); err != nil {
			slog.Debug("tcpdump exited", "link", linkID, "error", err)
		}
		m.mu.Lock()
		session.Active = false
		m.mu.Unlock()
	}()

	m.sessions[linkID] = session
	slog.Info("capture started", "link", linkID, "interface", iface, "file", filePath)
	return session, nil
}

// Stop stops an active capture on the given link.
func (m *CaptureManager) Stop(linkID string) error {
	m.mu.Lock()
	session, ok := m.sessions[linkID]
	if !ok || !session.Active {
		m.mu.Unlock()
		return fmt.Errorf("no active capture on link %s", linkID)
	}
	m.mu.Unlock()

	// Signal tcpdump to stop
	if err := m.executor.Stop(session.cmd); err != nil {
		session.cancel()
	}

	// Wait for process to finish
	select {
	case <-session.done:
	case <-time.After(5 * time.Second):
		session.cancel()
		<-session.done
	}

	slog.Info("capture stopped", "link", linkID, "file", session.FilePath)
	return nil
}

// GetSession returns the current capture session for a link, if any.
func (m *CaptureManager) GetSession(linkID string) *CaptureSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[linkID]
}

// GetFilePath returns the pcap file path for a link's capture.
func (m *CaptureManager) GetFilePath(linkID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[linkID]
	if !ok {
		return "", fmt.Errorf("no capture session for link %s", linkID)
	}
	return session.FilePath, nil
}

// Cleanup removes the pcap file for a link's capture.
func (m *CaptureManager) Cleanup(linkID string) error {
	m.mu.Lock()
	session, ok := m.sessions[linkID]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	if session.Active {
		m.mu.Unlock()
		return fmt.Errorf("cannot cleanup active capture on link %s", linkID)
	}
	delete(m.sessions, linkID)
	m.mu.Unlock()

	if err := os.Remove(session.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing pcap file: %w", err)
	}
	return nil
}

func sanitizeFileName(s string) string {
	var result []byte
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}
