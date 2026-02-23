package network

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
)

// VethOperator abstracts netlink operations for testing.
type VethOperator interface {
	LinkByName(name string) (netlink.Link, error)
	LinkSetUp(link netlink.Link) error
	LinkSetDown(link netlink.Link) error
	LinkList() ([]netlink.Link, error)
	QdiscAdd(qdisc netlink.Qdisc) error
	QdiscDel(qdisc netlink.Qdisc) error
}

// RealVethOperator uses real netlink operations.
type RealVethOperator struct{}

func (o *RealVethOperator) LinkByName(name string) (netlink.Link, error) {
	return netlink.LinkByName(name)
}

func (o *RealVethOperator) LinkSetUp(link netlink.Link) error {
	return netlink.LinkSetUp(link)
}

func (o *RealVethOperator) LinkSetDown(link netlink.Link) error {
	return netlink.LinkSetDown(link)
}

func (o *RealVethOperator) LinkList() ([]netlink.Link, error) {
	return netlink.LinkList()
}

func (o *RealVethOperator) QdiscAdd(qdisc netlink.Qdisc) error {
	return netlink.QdiscAdd(qdisc)
}

func (o *RealVethOperator) QdiscDel(qdisc netlink.Qdisc) error {
	return netlink.QdiscDel(qdisc)
}

// FindHostVeth finds the host-side veth name for a container interface.
// It reads /proc/{pid}/net/ifindex to find the peer interface index.
func FindHostVeth(containerPID int, ifName string) (string, error) {
	// Read the container's network interfaces
	ifIndexPath := fmt.Sprintf("/proc/%d/net/dev", containerPID)
	data, err := os.ReadFile(ifIndexPath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", ifIndexPath, err)
	}

	var containerIfIndex int
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, ifName+":") {
			// Parse the interface index from /proc/net/dev is unreliable
			// Instead use /sys/class/net/{if}/iflink
			break
		}
	}

	// Read the iflink to find the peer's index on the host
	iflinkPath := fmt.Sprintf("/proc/%d/root/sys/class/net/%s/iflink", containerPID, ifName)
	iflinkData, err := os.ReadFile(iflinkPath)
	if err != nil {
		return "", fmt.Errorf("reading iflink for %s: %w", ifName, err)
	}

	containerIfIndex, err = strconv.Atoi(strings.TrimSpace(string(iflinkData)))
	if err != nil {
		return "", fmt.Errorf("parsing iflink: %w", err)
	}

	// Find the host-side interface with this index
	links, err := netlink.LinkList()
	if err != nil {
		return "", fmt.Errorf("listing host links: %w", err)
	}

	for _, link := range links {
		if link.Attrs().Index == containerIfIndex {
			return link.Attrs().Name, nil
		}
	}

	return "", fmt.Errorf("host veth with index %d not found", containerIfIndex)
}
