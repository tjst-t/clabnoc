package network

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// NetemParams holds tc netem parameters.
type NetemParams struct {
	DelayMS          int    `json:"delay_ms"`
	JitterMS         int    `json:"jitter_ms"`
	LossPercent      int    `json:"loss_percent"`
	CorruptPercent   int    `json:"corrupt_percent"`
	DuplicatePercent int    `json:"duplicate_percent"`
	BPFFilter        string `json:"bpf_filter,omitempty"`
}

// FaultOperator abstracts fault injection operations on container interfaces.
type FaultOperator interface {
	LinkSetDown(ctx context.Context, containerID, ifName string) error
	LinkSetUp(ctx context.Context, containerID, ifName string) error
	ApplyNetem(ctx context.Context, containerID, ifName string, params *NetemParams) error
	ClearNetem(ctx context.Context, containerID, ifName string) error
}

// ExecFunc is a function that executes a command inside a container.
type ExecFunc func(ctx context.Context, containerID string, cmd []string) (string, error)

// DockerFaultOperator implements FaultOperator using docker exec.
type DockerFaultOperator struct {
	execFn ExecFunc
}

// NewDockerFaultOperator creates a new DockerFaultOperator.
func NewDockerFaultOperator(execFn ExecFunc) *DockerFaultOperator {
	return &DockerFaultOperator{execFn: execFn}
}

func (o *DockerFaultOperator) LinkSetDown(ctx context.Context, containerID, ifName string) error {
	out, err := o.execFn(ctx, containerID, []string{"ip", "link", "set", ifName, "down"})
	if err != nil {
		return fmt.Errorf("ip link set %s down: %w (output: %s)", ifName, err, strings.TrimSpace(out))
	}
	return nil
}

func (o *DockerFaultOperator) LinkSetUp(ctx context.Context, containerID, ifName string) error {
	out, err := o.execFn(ctx, containerID, []string{"ip", "link", "set", ifName, "up"})
	if err != nil {
		return fmt.Errorf("ip link set %s up: %w (output: %s)", ifName, err, strings.TrimSpace(out))
	}
	return nil
}

func (o *DockerFaultOperator) ApplyNetem(ctx context.Context, containerID, ifName string, params *NetemParams) error {
	if params.BPFFilter != "" {
		return o.applyFilteredNetem(ctx, containerID, ifName, params)
	}
	return o.applySimpleNetem(ctx, containerID, ifName, params)
}

func (o *DockerFaultOperator) applySimpleNetem(ctx context.Context, containerID, ifName string, params *NetemParams) error {
	// First try to delete existing qdisc (ignore error)
	_, _ = o.execFn(ctx, containerID, []string{"tc", "qdisc", "del", "dev", ifName, "root"})

	cmd := append([]string{"tc", "qdisc", "add", "dev", ifName, "root", "netem"}, buildNetemArgs(params)...)

	out, err := o.execFn(ctx, containerID, cmd)
	if err != nil {
		return fmt.Errorf("tc qdisc add netem on %s: %w (output: %s)", ifName, err, strings.TrimSpace(out))
	}
	return nil
}

// applyFilteredNetem sets up a prio qdisc with netem on band 1 and tc filter
// to classify only matching packets into the netem band.
// Structure: prio (root 1:0, 3 bands) -> netem (1:1) + passthrough (1:2, 1:3)
// tc filter classifies matching packets to flowid 1:1
func (o *DockerFaultOperator) applyFilteredNetem(ctx context.Context, containerID, ifName string, params *NetemParams) error {
	// 1. Delete existing root qdisc (ignore error)
	_, _ = o.execFn(ctx, containerID, []string{"tc", "qdisc", "del", "dev", ifName, "root"})

	// 2. Add prio root qdisc with 3 bands, all traffic defaults to band 3 (passthrough)
	out, err := o.execFn(ctx, containerID, []string{
		"tc", "qdisc", "add", "dev", ifName, "root", "handle", "1:", "prio", "bands", "3",
		"priomap", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2",
	})
	if err != nil {
		return fmt.Errorf("tc prio add on %s: %w (output: %s)", ifName, err, strings.TrimSpace(out))
	}

	// 3. Add netem child qdisc on band 1 (handle 1:1)
	netemCmd := append([]string{
		"tc", "qdisc", "add", "dev", ifName, "parent", "1:1", "handle", "10:", "netem",
	}, buildNetemArgs(params)...)

	out, err = o.execFn(ctx, containerID, netemCmd)
	if err != nil {
		return fmt.Errorf("tc netem add on %s: %w (output: %s)", ifName, err, strings.TrimSpace(out))
	}

	// 4. Add tc filter rules to classify matching packets to band 1
	filterCmds, err := BuildTCFilterCommands(ifName, params.BPFFilter)
	if err != nil {
		return fmt.Errorf("building tc filter for %q: %w", params.BPFFilter, err)
	}

	for _, filterArgs := range filterCmds {
		cmd := append([]string{"tc"}, filterArgs...)
		out, err = o.execFn(ctx, containerID, cmd)
		if err != nil {
			return fmt.Errorf("tc filter add on %s: %w (output: %s)", ifName, err, strings.TrimSpace(out))
		}
	}

	return nil
}

// buildNetemArgs builds the netem-specific arguments from params.
func buildNetemArgs(params *NetemParams) []string {
	var args []string
	if params.DelayMS > 0 {
		args = append(args, "delay", fmt.Sprintf("%dms", params.DelayMS))
		if params.JitterMS > 0 {
			args = append(args, fmt.Sprintf("%dms", params.JitterMS))
		}
	}
	if params.LossPercent > 0 {
		args = append(args, "loss", fmt.Sprintf("%d%%", params.LossPercent))
	}
	if params.CorruptPercent > 0 {
		args = append(args, "corrupt", fmt.Sprintf("%d%%", params.CorruptPercent))
	}
	if params.DuplicatePercent > 0 {
		args = append(args, "duplicate", fmt.Sprintf("%d%%", params.DuplicatePercent))
	}
	return args
}

func (o *DockerFaultOperator) ClearNetem(ctx context.Context, containerID, ifName string) error {
	out, err := o.execFn(ctx, containerID, []string{"tc", "qdisc", "del", "dev", ifName, "root"})
	if err != nil {
		return fmt.Errorf("tc qdisc del on %s: %w (output: %s)", ifName, err, strings.TrimSpace(out))
	}
	return nil
}

// FaultManager manages link fault injection state and operations.
type FaultManager struct {
	operator FaultOperator
	states   map[string]*LinkState
}

// NewFaultManager creates a new FaultManager.
func NewFaultManager(operator FaultOperator) *FaultManager {
	return &FaultManager{
		operator: operator,
		states:   make(map[string]*LinkState),
	}
}

// SetEndpointMapping registers the container endpoint targets for a link.
func (fm *FaultManager) SetEndpointMapping(linkID string, a, z *EndpointTarget) {
	state := fm.getOrCreate(linkID)
	state.A = a
	state.Z = z
}

// LinkDown brings down both endpoint interfaces for a link.
func (fm *FaultManager) LinkDown(ctx context.Context, linkID string) error {
	state := fm.getOrCreate(linkID)
	if state.A == nil && state.Z == nil {
		return fmt.Errorf("no endpoint mapping for link %s", linkID)
	}

	for _, ep := range []*EndpointTarget{state.A, state.Z} {
		if ep == nil {
			continue
		}
		if err := fm.operator.LinkSetDown(ctx, ep.ContainerID, ep.Interface); err != nil {
			return fmt.Errorf("setting %s down in %s: %w", ep.Interface, shortID(ep.ContainerID), err)
		}
	}

	state.State = "down"
	slog.Info("link set down", "link", linkID)
	return nil
}

// LinkUp brings up both endpoint interfaces for a link.
func (fm *FaultManager) LinkUp(ctx context.Context, linkID string) error {
	state := fm.getOrCreate(linkID)
	if state.A == nil && state.Z == nil {
		return fmt.Errorf("no endpoint mapping for link %s", linkID)
	}

	for _, ep := range []*EndpointTarget{state.A, state.Z} {
		if ep == nil {
			continue
		}
		if err := fm.operator.LinkSetUp(ctx, ep.ContainerID, ep.Interface); err != nil {
			return fmt.Errorf("setting %s up in %s: %w", ep.Interface, shortID(ep.ContainerID), err)
		}
	}

	state.State = "up"
	state.Netem = nil
	slog.Info("link set up", "link", linkID)
	return nil
}

// ApplyNetem applies tc netem to both endpoint interfaces.
func (fm *FaultManager) ApplyNetem(ctx context.Context, linkID string, params *NetemParams) error {
	state := fm.getOrCreate(linkID)
	if state.A == nil && state.Z == nil {
		return fmt.Errorf("no endpoint mapping for link %s", linkID)
	}

	for _, ep := range []*EndpointTarget{state.A, state.Z} {
		if ep == nil {
			continue
		}
		if err := fm.operator.ApplyNetem(ctx, ep.ContainerID, ep.Interface, params); err != nil {
			return fmt.Errorf("applying netem to %s in %s: %w", ep.Interface, shortID(ep.ContainerID), err)
		}
	}

	state.State = "degraded"
	state.Netem = params
	slog.Info("netem applied", "link", linkID, "delay_ms", params.DelayMS, "loss", params.LossPercent)
	return nil
}

// ClearNetem removes tc netem from both endpoint interfaces.
func (fm *FaultManager) ClearNetem(ctx context.Context, linkID string) error {
	state := fm.getOrCreate(linkID)
	if state.A == nil && state.Z == nil {
		return fmt.Errorf("no endpoint mapping for link %s", linkID)
	}

	for _, ep := range []*EndpointTarget{state.A, state.Z} {
		if ep == nil {
			continue
		}
		if err := fm.operator.ClearNetem(ctx, ep.ContainerID, ep.Interface); err != nil {
			slog.Warn("failed to clear netem", "container", shortID(ep.ContainerID), "interface", ep.Interface, "error", err)
		}
	}

	state.State = "up"
	state.Netem = nil
	slog.Info("netem cleared", "link", linkID)
	return nil
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func (fm *FaultManager) getOrCreate(linkID string) *LinkState {
	if s, ok := fm.states[linkID]; ok {
		return s
	}
	s := &LinkState{State: "up"}
	fm.states[linkID] = s
	return s
}
