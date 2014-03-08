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
	hostsseen := make(map[string]int)
	baseblock := &Block{"", swarm, "", nil, nil}
	for i := 0; i < len(hosts); i++ {
		// Create all the blockchains and let them do nothing.
		swarms = append(swarms, NewBlockchain(hosts[i], swarm, nil))
		if state, ok := swarms[i].state.(*StateSwarmInformed); ok {
			// Prevent race hazards
			state.Die(true)
		}
		hostsseen[hosts[i]] = 1
		swarms[i].AddSource(mult)
	}
	for i := 0; i < len(swarms); i++ {
		// Set all blockchains to StateSteady
		// The initial secret string needs to be unique, use host ID
		swarms[i].state = NewStateSteady(swarms[i], baseblock, hostsseen, hosts[i])
	}
	time.Sleep(5 * time.Second)
	// Die swarm, you don't belong in this world
	for _, v := range swarms {
		switch s := v.state.(type) {
		case *StateSteady:
			s.die <- struct{}{}
		default:
			t.Fatal("State unexpectedly changed")
		}
	}
}
