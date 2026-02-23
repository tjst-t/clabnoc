package network

import (
	"sync"
	"time"
)

// FaultState tracks the injected fault state of links.
type FaultState struct {
	mu     sync.RWMutex
	faults map[string]*LinkFaultState
}

// LinkFaultState holds the fault state for a single link.
type LinkFaultState struct {
	LinkID    string
	State     string // "down", "degraded"
	Netem     *NetemParams
	AppliedAt time.Time
	VethA     string // host veth name for endpoint A
	VethZ     string // host veth name for endpoint Z
}

// NewFaultState creates a new FaultState.
func NewFaultState() *FaultState {
	return &FaultState{faults: make(map[string]*LinkFaultState)}
}

// Set stores a link fault state.
func (s *FaultState) Set(linkID string, state *LinkFaultState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faults[linkID] = state
}

// Get retrieves the fault state for a link.
func (s *FaultState) Get(linkID string) (*LinkFaultState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.faults[linkID]
	return f, ok
}

// Delete removes the fault state for a link.
func (s *FaultState) Delete(linkID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.faults, linkID)
}

// GetLinkState returns the current fault state string for a link.
// Returns "up" if no fault state is recorded.
func (s *FaultState) GetLinkState(linkID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if f, ok := s.faults[linkID]; ok {
		return f.State
	}
	return "up"
}
