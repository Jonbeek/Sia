package swarm

import (
	"common"
	"log"
	"network"
	"testing"
	"time"
)

func TestStateJoin(t *testing.T) {

	log.SetFlags(log.Lmicroseconds)

	old := common.SWARMSIZE
	common.SWARMSIZE = 4
	defer func(old int) {
		common.SWARMSIZE = old
	}(old)

	mult := network.NewNetworkMultiplexer()

	hosts := make([]string, common.SWARMSIZE)

	storage := make(map[string]interface{})

	swarm := "test"

	for i, _ := range hosts {
		hosts[i], _ = common.RandomString(8)
		if len(hosts[i]) == 0 {
			t.Fatal(hosts)
		}
		storage[hosts[i]] = nil
	}

	swarms := make([]*Blockchain, common.SWARMSIZE)

	for i, _ := range swarms {
		swarms[i] = NewBlockchain(hosts[i], swarm, storage)
		if len(swarms[i].Host) == 0 {
			t.Fatal(swarms[i])
		}
	}

	for _, s := range swarms {
		s.AddSource(mult)
	}

	time.Sleep(3 * time.Second)
	log.Print("TEST: stopped sleeping")

	informed := 0
	broadcast := uint(0)
	seen := 0
	connected := 0
	blocks := 0

	for _, s := range swarms {
		switch t := s.state.(type) {
		case *StateSwarmInformed:
			t.Die(true)
			informed += 1
			broadcast += t.broadcastcount
			seen += len(t.hostsseen)
			blocks += len(t.chain.BlockHistory)
		case *StateSteady:
			connected += 1
		}
	}

	t.Log("StateInformed", informed)
	t.Log("BroadCasts Sent", broadcast)
	t.Log("PeersSeen", seen)
	t.Log("Blocks", blocks)
	t.Log("StateConnected", connected)

	t.Log(swarms[0].state)

	if connected <= common.SWARMSIZE/2 {
		t.Fatal("Failed to establish swarm")
	}
}
