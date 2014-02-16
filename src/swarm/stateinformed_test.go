package swarm

import (
	"common"
	"network"
	"testing"
	"time"
)

func TestStateJoin(t *testing.T) {
	mult := network.NewSimpleMultiplexer()

	hosts := make([]string, 256)

	storage := make(map[string]interface{})

	swarm := "test"

	for i, _ := range hosts {
		hosts[i], _ = common.RandomString(8)
		storage[hosts[i]] = nil
	}

	swarms := make([]*BlockChain, 256)

	for i, _ := range swarms {
		swarms[i] = NewBlockChain(hosts[i], swarm, storage)
	}

	for _, s := range swarms {
		s.AddSource(mult)
	}

	time.Sleep(4 * time.Second)

	informed := 0
	broadcast := uint(0)
	seen := 0
	connected := 0
	blocks := 0

	for _, s := range swarms {
		switch t := s.state.(type) {
		case *StateSwarmInformed:
			informed += 1
			broadcast += t.broadcastcount
			seen += len(t.hostsseen)
			blocks += len(t.chain.BlockHistory)
		case *StateSwarmConnected:
			connected += 1
		}
	}

	t.Log("StateInformed", informed)
	t.Log("BroadCasts Sent", broadcast)
	t.Log("PeersSeen", seen)
	t.Log("Blocks", blocks)
	t.Log("StateConnected", connected)

	t.Log(swarms[0].state)

	if connected <= 128 {
		t.Fatal("Failed to establish swarm")
	}
}
