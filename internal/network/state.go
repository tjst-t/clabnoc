package network

// EndpointTarget identifies a container interface for fault injection.
type EndpointTarget struct {
	ContainerID string `json:"container_id"`
	Interface   string `json:"interface"`
}

// LinkState holds the current fault injection state for a link.
type LinkState struct {
	State string       `json:"state"`
	Netem *NetemParams `json:"netem,omitempty"`
	A     *EndpointTarget `json:"a,omitempty"`
	Z     *EndpointTarget `json:"z,omitempty"`
}

// GetState returns the current state for a link.
func (fm *FaultManager) GetState(linkID string) LinkState {
	if s, ok := fm.states[linkID]; ok {
		return *s
	}
	return LinkState{State: "up"}
}
