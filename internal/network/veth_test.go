package network_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tjst-t/clabnoc/internal/network"
)

func TestFaultManager_New(t *testing.T) {
	fm := network.NewFaultManager()
	assert.NotNil(t, fm)
}

func TestFaultState_SetAndGet(t *testing.T) {
	fs := network.NewFaultState()
	state := &network.LinkFaultState{
		LinkID: "node1:eth1__node2:eth1",
		State:  "down",
		VethA:  "veth123",
		VethZ:  "veth456",
	}
	fs.Set("node1:eth1__node2:eth1", state)
	got, ok := fs.Get("node1:eth1__node2:eth1")
	assert.True(t, ok)
	assert.Equal(t, "down", got.State)
	assert.Equal(t, "veth123", got.VethA)
	assert.Equal(t, "veth456", got.VethZ)
}

func TestFaultState_GetDefault(t *testing.T) {
	fs := network.NewFaultState()
	state := fs.GetLinkState("nonexistent")
	assert.Equal(t, "up", state)
}

func TestFaultState_Delete(t *testing.T) {
	fs := network.NewFaultState()
	fs.Set("link1", &network.LinkFaultState{LinkID: "link1", State: "down"})
	fs.Delete("link1")
	_, ok := fs.Get("link1")
	assert.False(t, ok)
}

func TestFaultState_GetLinkState_Down(t *testing.T) {
	fs := network.NewFaultState()
	fs.Set("link1", &network.LinkFaultState{LinkID: "link1", State: "down"})
	state := fs.GetLinkState("link1")
	assert.Equal(t, "down", state)
}

func TestFaultState_GetLinkState_Degraded(t *testing.T) {
	fs := network.NewFaultState()
	fs.Set("link2", &network.LinkFaultState{LinkID: "link2", State: "degraded"})
	state := fs.GetLinkState("link2")
	assert.Equal(t, "degraded", state)
}

func TestFaultState_GetLinkState_AfterDelete(t *testing.T) {
	fs := network.NewFaultState()
	fs.Set("link1", &network.LinkFaultState{LinkID: "link1", State: "down"})
	fs.Delete("link1")
	state := fs.GetLinkState("link1")
	assert.Equal(t, "up", state)
}

func TestFaultState_WithNetem(t *testing.T) {
	fs := network.NewFaultState()
	params := &network.NetemParams{
		DelayMs:     100,
		JitterMs:    10,
		LossPercent: 1.5,
	}
	fs.Set("link1", &network.LinkFaultState{
		LinkID:    "link1",
		State:     "degraded",
		Netem:     params,
		AppliedAt: time.Now(),
	})
	got, ok := fs.Get("link1")
	assert.True(t, ok)
	assert.Equal(t, "degraded", got.State)
	assert.NotNil(t, got.Netem)
	assert.Equal(t, 100, got.Netem.DelayMs)
	assert.Equal(t, 1.5, got.Netem.LossPercent)
}

func TestNetemParams_Structure(t *testing.T) {
	p := network.NetemParams{
		DelayMs:          100,
		JitterMs:         10,
		LossPercent:      30.0,
		CorruptPercent:   0,
		DuplicatePercent: 0,
	}
	assert.Equal(t, 100, p.DelayMs)
	assert.Equal(t, 30.0, p.LossPercent)
}

func TestFaultState_ConcurrentAccess(t *testing.T) {
	fs := network.NewFaultState()
	done := make(chan struct{})

	// Write concurrently
	go func() {
		for i := 0; i < 100; i++ {
			fs.Set("link1", &network.LinkFaultState{LinkID: "link1", State: "down"})
		}
		close(done)
	}()

	// Read concurrently
	for i := 0; i < 100; i++ {
		fs.GetLinkState("link1")
	}

	<-done
}

func TestVethInfo_Structure(t *testing.T) {
	vi := network.VethInfo{
		Name:      "veth0abc",
		Index:     42,
		PeerIndex: 43,
		State:     "up",
	}
	assert.Equal(t, "veth0abc", vi.Name)
	assert.Equal(t, 42, vi.Index)
	assert.Equal(t, 43, vi.PeerIndex)
	assert.Equal(t, "up", vi.State)
}

func TestNetemParams_LossPercent(t *testing.T) {
	p := network.NetemParams{LossPercent: 30.0}
	assert.Equal(t, 30.0, p.LossPercent)
	assert.True(t, p.LossPercent >= 0 && p.LossPercent <= 100)
}

func TestNetemParams_DelayMs(t *testing.T) {
	p := network.NetemParams{DelayMs: 100, JitterMs: 10}
	assert.Equal(t, 100, p.DelayMs)
	assert.Equal(t, 10, p.JitterMs)
}

func TestNetemParams_ZeroValues(t *testing.T) {
	p := network.NetemParams{}
	assert.Equal(t, 0, p.DelayMs)
	assert.Equal(t, 0, p.JitterMs)
	assert.Equal(t, 0.0, p.LossPercent)
	assert.Equal(t, 0.0, p.CorruptPercent)
	assert.Equal(t, 0.0, p.DuplicatePercent)
}

func TestNetemParams_AllFields(t *testing.T) {
	p := network.NetemParams{
		DelayMs:          50,
		JitterMs:         5,
		LossPercent:      1.5,
		CorruptPercent:   0.1,
		DuplicatePercent: 0.5,
	}
	assert.Equal(t, 50, p.DelayMs)
	assert.Equal(t, 5, p.JitterMs)
	assert.Equal(t, 1.5, p.LossPercent)
	assert.Equal(t, 0.1, p.CorruptPercent)
	assert.Equal(t, 0.5, p.DuplicatePercent)
}
