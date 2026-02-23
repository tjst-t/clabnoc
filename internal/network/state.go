package network

// LinkState holds the current fault injection state for a link.
type LinkState struct {
	State     string       `json:"state"`
	Netem     *NetemParams `json:"netem,omitempty"`
	HostVethA string       `json:"host_veth_a,omitempty"`
	HostVethZ string       `json:"host_veth_z,omitempty"`
}

// GetState returns the current state for a link.
func (fm *FaultManager) GetState(linkID string) LinkState {
	if s, ok := fm.states[linkID]; ok {
		return *s
	}
	return LinkState{State: "up"}
}
