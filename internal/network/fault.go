package network

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

// FaultManager manages link fault state.
type FaultManager struct {
	// Track which links have netem applied
	netemLinks map[string]bool
}

// NewFaultManager creates a new FaultManager.
func NewFaultManager() *FaultManager {
	return &FaultManager{netemLinks: make(map[string]bool)}
}

// LinkDown brings down a veth interface (fault injection).
func (f *FaultManager) LinkDown(ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("link %s not found: %w", ifName, err)
	}
	if err := netlink.LinkSetDown(link); err != nil {
		return fmt.Errorf("link down %s: %w", ifName, err)
	}
	return nil
}

// LinkUp brings up a veth interface (fault recovery).
func (f *FaultManager) LinkUp(ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("link %s not found: %w", ifName, err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("link up %s: %w", ifName, err)
	}
	return nil
}

// ApplyNetem applies tc netem to a veth interface.
func (f *FaultManager) ApplyNetem(ifName string, cfg NetemParams) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("link %s not found: %w", ifName, err)
	}

	// Remove existing qdisc first
	_ = f.ClearNetem(ifName)

	attrs := netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	}
	netemAttrs := netlink.NetemQdiscAttrs{
		Latency:     uint32(cfg.DelayMs * 1000),        // microseconds
		Jitter:      uint32(cfg.JitterMs * 1000),        // microseconds
		Loss:        float32(cfg.LossPercent),           // in %
		Duplicate:   float32(cfg.DuplicatePercent),      // in %
		CorruptProb: float32(cfg.CorruptPercent),        // in %
	}
	qdisc := netlink.NewNetem(attrs, netemAttrs)

	if err := netlink.QdiscAdd(qdisc); err != nil {
		return fmt.Errorf("apply netem on %s: %w", ifName, err)
	}
	f.netemLinks[ifName] = true
	return nil
}

// ClearNetem removes tc netem from a veth interface.
func (f *FaultManager) ClearNetem(ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("link %s not found: %w", ifName, err)
	}

	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return fmt.Errorf("list qdiscs %s: %w", ifName, err)
	}
	for _, q := range qdiscs {
		if q.Type() == "netem" {
			if err := netlink.QdiscDel(q); err != nil {
				return fmt.Errorf("del qdisc %s: %w", ifName, err)
			}
		}
	}
	delete(f.netemLinks, ifName)
	return nil
}

// GetLinkState returns the current operational state of a link.
func GetLinkState(ifName string) (string, error) {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return "", fmt.Errorf("link %s not found: %w", ifName, err)
	}
	attrs := link.Attrs()
	if attrs.Flags&1 == 0 { // net.FlagUp
		return "down", nil
	}
	return "up", nil
}

// NetemParams holds network emulation configuration.
type NetemParams struct {
	DelayMs          int
	JitterMs         int
	LossPercent      float64
	CorruptPercent   float64
	DuplicatePercent float64
}
