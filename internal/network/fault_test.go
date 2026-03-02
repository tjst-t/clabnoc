package network

import (
	"context"
	"fmt"
	"testing"
)

// execCall records a docker exec call for test verification.
type execCall struct {
	ContainerID string
	Interface   string
	Action      string
}

// mockFaultOperator implements FaultOperator for testing.
type mockFaultOperator struct {
	calls   []execCall
	downErr error
	upErr   error
	addErr  error
	delErr  error
}

func (m *mockFaultOperator) LinkSetDown(ctx context.Context, containerID, ifName string) error {
	m.calls = append(m.calls, execCall{containerID, ifName, "down"})
	return m.downErr
}

func (m *mockFaultOperator) LinkSetUp(ctx context.Context, containerID, ifName string) error {
	m.calls = append(m.calls, execCall{containerID, ifName, "up"})
	return m.upErr
}

func (m *mockFaultOperator) ApplyNetem(ctx context.Context, containerID, ifName string, params *NetemParams) error {
	m.calls = append(m.calls, execCall{containerID, ifName, "netem"})
	return m.addErr
}

func (m *mockFaultOperator) ClearNetem(ctx context.Context, containerID, ifName string) error {
	m.calls = append(m.calls, execCall{containerID, ifName, "clear_netem"})
	return m.delErr
}

func TestLinkDown(t *testing.T) {
	op := &mockFaultOperator{}
	fm := NewFaultManager(op)
	fm.SetEndpointMapping("link1",
		&EndpointTarget{ContainerID: "container-aaa", Interface: "eth1"},
		&EndpointTarget{ContainerID: "container-zzz", Interface: "eth2"},
	)

	err := fm.LinkDown(context.Background(), "link1")
	if err != nil {
		t.Fatalf("LinkDown failed: %v", err)
	}

	if len(op.calls) != 2 {
		t.Fatalf("expected 2 down calls, got %d", len(op.calls))
	}
	if op.calls[0].ContainerID != "container-aaa" || op.calls[0].Interface != "eth1" || op.calls[0].Action != "down" {
		t.Errorf("unexpected call[0]: %+v", op.calls[0])
	}
	if op.calls[1].ContainerID != "container-zzz" || op.calls[1].Interface != "eth2" || op.calls[1].Action != "down" {
		t.Errorf("unexpected call[1]: %+v", op.calls[1])
	}

	state := fm.GetState("link1")
	if state.State != "down" {
		t.Errorf("expected state down, got %s", state.State)
	}
}

func TestLinkUp(t *testing.T) {
	op := &mockFaultOperator{}
	fm := NewFaultManager(op)
	fm.SetEndpointMapping("link1",
		&EndpointTarget{ContainerID: "container-aaa", Interface: "eth1"},
		&EndpointTarget{ContainerID: "container-zzz", Interface: "eth2"},
	)
	_ = fm.LinkDown(context.Background(), "link1")
	op.calls = nil // reset

	err := fm.LinkUp(context.Background(), "link1")
	if err != nil {
		t.Fatalf("LinkUp failed: %v", err)
	}

	if len(op.calls) != 2 {
		t.Fatalf("expected 2 up calls, got %d", len(op.calls))
	}
	for _, c := range op.calls {
		if c.Action != "up" {
			t.Errorf("expected action up, got %s", c.Action)
		}
	}

	state := fm.GetState("link1")
	if state.State != "up" {
		t.Errorf("expected state up, got %s", state.State)
	}
}

func TestLinkDownNoMapping(t *testing.T) {
	op := &mockFaultOperator{}
	fm := NewFaultManager(op)

	err := fm.LinkDown(context.Background(), "unknown")
	if err == nil {
		t.Error("expected error for unknown link")
	}
}

func TestApplyNetem(t *testing.T) {
	op := &mockFaultOperator{}
	fm := NewFaultManager(op)
	fm.SetEndpointMapping("link1",
		&EndpointTarget{ContainerID: "container-aaa", Interface: "eth1"},
		&EndpointTarget{ContainerID: "container-zzz", Interface: "eth2"},
	)

	params := &NetemParams{
		DelayMS:     100,
		JitterMS:    10,
		LossPercent: 30,
	}

	err := fm.ApplyNetem(context.Background(), "link1", params)
	if err != nil {
		t.Fatalf("ApplyNetem failed: %v", err)
	}

	if len(op.calls) != 2 {
		t.Fatalf("expected 2 netem calls, got %d", len(op.calls))
	}
	for _, c := range op.calls {
		if c.Action != "netem" {
			t.Errorf("expected action netem, got %s", c.Action)
		}
	}

	state := fm.GetState("link1")
	if state.State != "degraded" {
		t.Errorf("expected state degraded, got %s", state.State)
	}
	if state.Netem == nil {
		t.Fatal("expected netem params")
	}
	if state.Netem.DelayMS != 100 {
		t.Errorf("expected delay 100ms, got %d", state.Netem.DelayMS)
	}
}

func TestClearNetem(t *testing.T) {
	op := &mockFaultOperator{}
	fm := NewFaultManager(op)
	fm.SetEndpointMapping("link1",
		&EndpointTarget{ContainerID: "container-aaa", Interface: "eth1"},
		&EndpointTarget{ContainerID: "container-zzz", Interface: "eth2"},
	)
	_ = fm.ApplyNetem(context.Background(), "link1", &NetemParams{DelayMS: 100})
	op.calls = nil

	err := fm.ClearNetem(context.Background(), "link1")
	if err != nil {
		t.Fatalf("ClearNetem failed: %v", err)
	}

	if len(op.calls) != 2 {
		t.Fatalf("expected 2 clear calls, got %d", len(op.calls))
	}

	state := fm.GetState("link1")
	if state.State != "up" {
		t.Errorf("expected state up, got %s", state.State)
	}
	if state.Netem != nil {
		t.Error("expected nil netem")
	}
}

func TestGetStateDefault(t *testing.T) {
	op := &mockFaultOperator{}
	fm := NewFaultManager(op)

	state := fm.GetState("nonexistent")
	if state.State != "up" {
		t.Errorf("expected default state up, got %s", state.State)
	}
}

func TestLinkDownError(t *testing.T) {
	op := &mockFaultOperator{downErr: fmt.Errorf("exec failed")}
	fm := NewFaultManager(op)
	fm.SetEndpointMapping("link1",
		&EndpointTarget{ContainerID: "container-aaa", Interface: "eth1"},
		nil,
	)

	err := fm.LinkDown(context.Background(), "link1")
	if err == nil {
		t.Error("expected error")
	}
}

// Tests for DockerFaultOperator (exec-level behavior)

type rawExecCall struct {
	ContainerID string
	Cmd         []string
}

func TestDockerFaultOperatorSimpleNetem(t *testing.T) {
	var calls []rawExecCall
	execFn := func(ctx context.Context, containerID string, cmd []string) (string, error) {
		calls = append(calls, rawExecCall{containerID, cmd})
		return "", nil
	}

	op := NewDockerFaultOperator(execFn)
	params := &NetemParams{DelayMS: 100, LossPercent: 5}

	err := op.ApplyNetem(context.Background(), "ctr1", "eth0", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 calls: delete old qdisc + add netem
	if len(calls) != 2 {
		t.Fatalf("expected 2 exec calls, got %d", len(calls))
	}

	// First call: delete existing root qdisc
	if calls[0].Cmd[0] != "tc" || calls[0].Cmd[3] != "dev" || calls[0].Cmd[5] != "root" {
		t.Errorf("unexpected delete cmd: %v", calls[0].Cmd)
	}

	// Second call: add netem
	cmd := calls[1].Cmd
	if cmd[0] != "tc" || cmd[5] != "root" || cmd[6] != "netem" {
		t.Errorf("unexpected netem cmd: %v", cmd)
	}
}

func TestDockerFaultOperatorFilteredNetem(t *testing.T) {
	var calls []rawExecCall
	execFn := func(ctx context.Context, containerID string, cmd []string) (string, error) {
		calls = append(calls, rawExecCall{containerID, cmd})
		return "", nil
	}

	op := NewDockerFaultOperator(execFn)
	params := &NetemParams{
		DelayMS:   200,
		BPFFilter: "tcp port 179",
	}

	err := op.ApplyNetem(context.Background(), "ctr1", "eth0", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected calls: (1) delete old qdisc, (2) add prio, (3) add netem child, (4,5) add filter rules (src+dst)
	if len(calls) != 5 {
		t.Fatalf("expected 5 exec calls for filtered netem, got %d", len(calls))
		for i, c := range calls {
			t.Logf("  call[%d]: %v", i, c.Cmd)
		}
	}

	// Call 0: delete root qdisc
	if calls[0].Cmd[2] != "del" {
		t.Errorf("call[0]: expected 'del', got cmd: %v", calls[0].Cmd)
	}

	// Call 1: add prio root qdisc
	found := false
	for _, arg := range calls[1].Cmd {
		if arg == "prio" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("call[1]: expected prio qdisc, got cmd: %v", calls[1].Cmd)
	}

	// Call 2: add netem child on parent 1:1
	foundNetem := false
	foundParent := false
	for i, arg := range calls[2].Cmd {
		if arg == "netem" {
			foundNetem = true
		}
		if arg == "parent" && i+1 < len(calls[2].Cmd) && calls[2].Cmd[i+1] == "1:1" {
			foundParent = true
		}
	}
	if !foundNetem {
		t.Errorf("call[2]: expected netem, got cmd: %v", calls[2].Cmd)
	}
	if !foundParent {
		t.Errorf("call[2]: expected parent 1:1, got cmd: %v", calls[2].Cmd)
	}

	// Calls 3,4: tc filter add
	for i := 3; i < 5; i++ {
		if calls[i].Cmd[0] != "tc" || calls[i].Cmd[1] != "filter" {
			t.Errorf("call[%d]: expected tc filter, got cmd: %v", i, calls[i].Cmd)
		}
	}
}

func TestDockerFaultOperatorFilteredNetemICMP(t *testing.T) {
	var calls []rawExecCall
	execFn := func(ctx context.Context, containerID string, cmd []string) (string, error) {
		calls = append(calls, rawExecCall{containerID, cmd})
		return "", nil
	}

	op := NewDockerFaultOperator(execFn)
	params := &NetemParams{
		LossPercent: 50,
		BPFFilter:   "icmp",
	}

	err := op.ApplyNetem(context.Background(), "ctr1", "eth0", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: delete + prio + netem + 1 filter rule (icmp)
	if len(calls) != 4 {
		t.Fatalf("expected 4 exec calls for ICMP filter, got %d", len(calls))
	}
}

func TestDockerFaultOperatorFilteredNetemInvalidFilter(t *testing.T) {
	execFn := func(ctx context.Context, containerID string, cmd []string) (string, error) {
		return "", nil
	}

	op := NewDockerFaultOperator(execFn)
	params := &NetemParams{
		DelayMS:   100,
		BPFFilter: "unsupported expression",
	}

	err := op.ApplyNetem(context.Background(), "ctr1", "eth0", params)
	if err == nil {
		t.Error("expected error for invalid BPF filter")
	}
}

func TestBuildNetemArgs(t *testing.T) {
	tests := []struct {
		name   string
		params *NetemParams
		want   int // number of args
	}{
		{"empty", &NetemParams{}, 0},
		{"delay only", &NetemParams{DelayMS: 100}, 2},
		{"delay with jitter", &NetemParams{DelayMS: 100, JitterMS: 10}, 3},
		{"loss only", &NetemParams{LossPercent: 5}, 2},
		{"all params", &NetemParams{DelayMS: 100, JitterMS: 10, LossPercent: 5, CorruptPercent: 3, DuplicatePercent: 2}, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildNetemArgs(tt.params)
			if len(args) != tt.want {
				t.Errorf("expected %d args, got %d: %v", tt.want, len(args), args)
			}
		})
	}
}
