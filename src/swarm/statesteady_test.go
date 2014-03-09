package swarm

import (
	"network"
	"testing"
	"time"
)

func TestStateSteady(t *testing.T) {
	mult := network.NewNetworkMultiplexer()
	hosts := []string{"A", "B", "C", "D"}
	swarm := "test"
	var swarms []*Blockchain
	hostsseen := make(map[string]*Heartbeat)
	baseblock := &Block{"", swarm, nil, nil}
	for i := 0; i < len(hosts); i++ {
		// Create all the blockchains and let them do nothing.
		swarms = append(swarms, newBlockchain(hosts[i], swarm, nil))
		hostsseen[hosts[i]] = nil
	}
	for i := 0; i < len(swarms); i++ {
		// Set all blockchains to StateSteady
		// The initial secret string needs to be unique, use host ID
		swarms[i].state = NewStateSteady(swarms[i], baseblock, hostsseen, swarms[i].Host)
		swarms[i].AddSource(mult)
	}
	time.Sleep(10 * time.Millisecond)
	// Die swarm, you don't belong in this world
	for _, v := range swarms {
		switch s := v.GetState().(type) {
		case *StateSteady:
			go s.Die()
		default:
			t.Fatal("State unexpectedly changed")
		}
	}

}
