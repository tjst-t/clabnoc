package network

import (
	"fmt"
	"log/slog"

	"github.com/vishvananda/netlink"
)

// NetemParams holds tc netem parameters.
type NetemParams struct {
	DelayMS          int `json:"delay_ms"`
	JitterMS         int `json:"jitter_ms"`
	LossPercent      int `json:"loss_percent"`
	CorruptPercent   int `json:"corrupt_percent"`
	DuplicatePercent int `json:"duplicate_percent"`
}

// FaultManager manages link fault injection state and operations.
type FaultManager struct {
	operator VethOperator
	states   map[string]*LinkState
}

// NewFaultManager creates a new FaultManager.
func NewFaultManager(operator VethOperator) *FaultManager {
	return &FaultManager{
		operator: operator,
		states:   make(map[string]*LinkState),
	}
}

// SetVethMapping registers the host veth names for a link.
func (fm *FaultManager) SetVethMapping(linkID, vethA, vethZ string) {
	state := fm.getOrCreate(linkID)
	state.HostVethA = vethA
	state.HostVethZ = vethZ
}

// LinkDown brings down both veth interfaces for a link.
func (fm *FaultManager) LinkDown(linkID string) error {
	state := fm.getOrCreate(linkID)
	if state.HostVethA == "" && state.HostVethZ == "" {
		return fmt.Errorf("no veth mapping for link %s", linkID)
	}

	for _, vethName := range []string{state.HostVethA, state.HostVethZ} {
		if vethName == "" {
			continue
		}
		link, err := fm.operator.LinkByName(vethName)
		if err != nil {
			return fmt.Errorf("finding veth %s: %w", vethName, err)
		}
		if err := fm.operator.LinkSetDown(link); err != nil {
			return fmt.Errorf("setting %s down: %w", vethName, err)
		}
	}

	state.State = "down"
	slog.Info("link set down", "link", linkID)
	return nil
}

// LinkUp brings up both veth interfaces for a link.
func (fm *FaultManager) LinkUp(linkID string) error {
	state := fm.getOrCreate(linkID)
	if state.HostVethA == "" && state.HostVethZ == "" {
		return fmt.Errorf("no veth mapping for link %s", linkID)
	}

	for _, vethName := range []string{state.HostVethA, state.HostVethZ} {
		if vethName == "" {
			continue
		}
		link, err := fm.operator.LinkByName(vethName)
		if err != nil {
			return fmt.Errorf("finding veth %s: %w", vethName, err)
		}
		if err := fm.operator.LinkSetUp(link); err != nil {
			return fmt.Errorf("setting %s up: %w", vethName, err)
		}
	}

	state.State = "up"
	state.Netem = nil
	slog.Info("link set up", "link", linkID)
	return nil
}

// ApplyNetem applies tc netem to both veth interfaces.
func (fm *FaultManager) ApplyNetem(linkID string, params *NetemParams) error {
	state := fm.getOrCreate(linkID)
	if state.HostVethA == "" && state.HostVethZ == "" {
		return fmt.Errorf("no veth mapping for link %s", linkID)
	}

	for _, vethName := range []string{state.HostVethA, state.HostVethZ} {
		if vethName == "" {
			continue
		}
		if err := fm.applyNetemToInterface(vethName, params); err != nil {
			return fmt.Errorf("applying netem to %s: %w", vethName, err)
		}
	}

	state.State = "degraded"
	state.Netem = params
	slog.Info("netem applied", "link", linkID, "delay_ms", params.DelayMS, "loss", params.LossPercent)
	return nil
}

// ClearNetem removes tc netem from both veth interfaces.
func (fm *FaultManager) ClearNetem(linkID string) error {
	state := fm.getOrCreate(linkID)
	if state.HostVethA == "" && state.HostVethZ == "" {
		return fmt.Errorf("no veth mapping for link %s", linkID)
	}

	for _, vethName := range []string{state.HostVethA, state.HostVethZ} {
		if vethName == "" {
			continue
		}
		if err := fm.clearNetemFromInterface(vethName); err != nil {
			slog.Warn("failed to clear netem", "veth", vethName, "error", err)
		}
	}

	state.State = "up"
	state.Netem = nil
	slog.Info("netem cleared", "link", linkID)
	return nil
}

func (fm *FaultManager) applyNetemToInterface(vethName string, params *NetemParams) error {
	link, err := fm.operator.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf("finding veth %s: %w", vethName, err)
	}

	// First try to delete existing qdisc
	_ = fm.clearNetemFromInterface(vethName)

	qdisc := &netlink.Netem{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(1, 0),
			Parent:    netlink.HANDLE_ROOT,
		},
		Latency:   uint32(params.DelayMS * 1000),   // convert ms to us
		Jitter:    uint32(params.JitterMS * 1000),
		Loss:      uint32(params.LossPercent),
		CorruptProb: uint32(params.CorruptPercent),
		Duplicate: uint32(params.DuplicatePercent),
	}

	return fm.operator.QdiscAdd(qdisc)
}

func (fm *FaultManager) clearNetemFromInterface(vethName string) error {
	link, err := fm.operator.LinkByName(vethName)
	if err != nil {
		return err
	}

	qdisc := &netlink.Netem{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(1, 0),
			Parent:    netlink.HANDLE_ROOT,
		},
	}

	return fm.operator.QdiscDel(qdisc)
}

func (fm *FaultManager) getOrCreate(linkID string) *LinkState {
	if s, ok := fm.states[linkID]; ok {
		return s
	}
	s := &LinkState{State: "up"}
	fm.states[linkID] = s
	return s
}
