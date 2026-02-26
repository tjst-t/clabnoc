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
