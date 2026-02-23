package network

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

// FindHostVethByPeerIndex finds the veth interface on the host that has the given peer index.
// Used to find the host-side veth corresponding to a container interface.
func FindHostVethByPeerIndex(peerIndex int) (string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return "", fmt.Errorf("list links: %w", err)
	}
	for _, link := range links {
		if veth, ok := link.(*netlink.Veth); ok {
			peerIdx, err := netlink.VethPeerIndex(veth)
			if err != nil {
				continue
			}
			if peerIdx == peerIndex {
				return veth.Name, nil
			}
		}
	}
	return "", fmt.Errorf("no veth found with peer index %d", peerIndex)
}

// GetLinkIndex returns the interface index for a given interface name.
func GetLinkIndex(name string) (int, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return 0, fmt.Errorf("link %s: %w", name, err)
	}
	return link.Attrs().Index, nil
}

// ListVeths returns all veth interfaces on the host.
func ListVeths() ([]VethInfo, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	var veths []VethInfo
	for _, link := range links {
		if veth, ok := link.(*netlink.Veth); ok {
			peerIdx, _ := netlink.VethPeerIndex(veth)
			veths = append(veths, VethInfo{
				Name:      veth.Name,
				Index:     veth.Index,
				PeerIndex: peerIdx,
				State:     link.Attrs().Flags.String(),
			})
		}
	}
	return veths, nil
}

// VethInfo holds information about a veth interface.
type VethInfo struct {
	Name      string
	Index     int
	PeerIndex int
	State     string
}
