package network

import (
	"fmt"
	"testing"

	"github.com/vishvananda/netlink"
)

// mockLink implements netlink.Link for testing.
type mockLink struct {
	attrs netlink.LinkAttrs
}

func (m *mockLink) Attrs() *netlink.LinkAttrs { return &m.attrs }
func (m *mockLink) Type() string              { return "veth" }

// mockVethOp implements VethOperator for testing.
type mockVethOp struct {
	links     map[string]netlink.Link
	downCalls []string
	upCalls   []string
	addCalls  int
	delCalls  int
	downErr   error
	upErr     error
	addErr    error
	delErr    error
}

func newMockVethOp() *mockVethOp {
	return &mockVethOp{
		links: make(map[string]netlink.Link),
	}
}

func (m *mockVethOp) addLink(name string, index int) {
	m.links[name] = &mockLink{attrs: netlink.LinkAttrs{Name: name, Index: index}}
}

func (m *mockVethOp) LinkByName(name string) (netlink.Link, error) {
	link, ok := m.links[name]
	if !ok {
		return nil, fmt.Errorf("link %s not found", name)
	}
	return link, nil
}

func (m *mockVethOp) LinkSetUp(link netlink.Link) error {
	m.upCalls = append(m.upCalls, link.Attrs().Name)
	return m.upErr
}

func (m *mockVethOp) LinkSetDown(link netlink.Link) error {
	m.downCalls = append(m.downCalls, link.Attrs().Name)
	return m.downErr
}

func (m *mockVethOp) LinkList() ([]netlink.Link, error) {
	links := make([]netlink.Link, 0, len(m.links))
	for _, l := range m.links {
		links = append(links, l)
	}
	return links, nil
}

func (m *mockVethOp) QdiscAdd(qdisc netlink.Qdisc) error {
	m.addCalls++
	return m.addErr
}

func (m *mockVethOp) QdiscDel(qdisc netlink.Qdisc) error {
	m.delCalls++
	return m.delErr
}

func TestLinkDown(t *testing.T) {
	op := newMockVethOp()
	op.addLink("veth-a", 10)
	op.addLink("veth-z", 11)

	fm := NewFaultManager(op)
	fm.SetVethMapping("link1", "veth-a", "veth-z")

	err := fm.LinkDown("link1")
	if err != nil {
		t.Fatalf("LinkDown failed: %v", err)
	}

	if len(op.downCalls) != 2 {
		t.Errorf("expected 2 down calls, got %d", len(op.downCalls))
	}

	state := fm.GetState("link1")
	if state.State != "down" {
		t.Errorf("expected state down, got %s", state.State)
	}
}

func TestLinkUp(t *testing.T) {
	op := newMockVethOp()
	op.addLink("veth-a", 10)
	op.addLink("veth-z", 11)

	fm := NewFaultManager(op)
	fm.SetVethMapping("link1", "veth-a", "veth-z")
	fm.LinkDown("link1") // Set to down first

	err := fm.LinkUp("link1")
	if err != nil {
		t.Fatalf("LinkUp failed: %v", err)
	}

	if len(op.upCalls) != 2 {
		t.Errorf("expected 2 up calls, got %d", len(op.upCalls))
	}

	state := fm.GetState("link1")
	if state.State != "up" {
		t.Errorf("expected state up, got %s", state.State)
	}
}

func TestLinkDownNoMapping(t *testing.T) {
	op := newMockVethOp()
	fm := NewFaultManager(op)

	err := fm.LinkDown("unknown")
	if err == nil {
		t.Error("expected error for unknown link")
	}
}

func TestApplyNetem(t *testing.T) {
	op := newMockVethOp()
	op.addLink("veth-a", 10)
	op.addLink("veth-z", 11)

	fm := NewFaultManager(op)
	fm.SetVethMapping("link1", "veth-a", "veth-z")

	params := &NetemParams{
		DelayMS:     100,
		JitterMS:    10,
		LossPercent: 30,
	}

	err := fm.ApplyNetem("link1", params)
	if err != nil {
		t.Fatalf("ApplyNetem failed: %v", err)
	}

	// 2 del (clear first) + 2 add = total
	if op.addCalls != 2 {
		t.Errorf("expected 2 add calls, got %d", op.addCalls)
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
	op := newMockVethOp()
	op.addLink("veth-a", 10)
	op.addLink("veth-z", 11)

	fm := NewFaultManager(op)
	fm.SetVethMapping("link1", "veth-a", "veth-z")
	fm.ApplyNetem("link1", &NetemParams{DelayMS: 100})

	err := fm.ClearNetem("link1")
	if err != nil {
		t.Fatalf("ClearNetem failed: %v", err)
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
	op := newMockVethOp()
	fm := NewFaultManager(op)

	state := fm.GetState("nonexistent")
	if state.State != "up" {
		t.Errorf("expected default state up, got %s", state.State)
	}
}
