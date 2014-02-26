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
	var hostsseen map[string]int
	baseblock := &Block{"", swarm, "", nil, nil}
	for i, v := range hosts {
		swarms[i] = NewBlockchain(v, swarm, nil)
		hostsseen[v] = 1
		swarms[i].AddSource(mult)
	}
	for i, v := range swarms {
		baseblock.Blockchain = hosts[i]
		v.state = NewStateSteady(v, baseblock, hostsseen)
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
