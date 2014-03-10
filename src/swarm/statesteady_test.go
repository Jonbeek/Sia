package swarm

import (
	"common"
	"log"
	"network"
	"testing"
	"time"
)

func TestStateSteady(t *testing.T) {

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

	start := time.Now().Add(100 * time.Millisecond)

	for i, _ := range swarms {
		swarms[i] = NewBlockchain(hosts[i], swarm, start, storage)
		if len(swarms[i].Host) == 0 {
			t.Fatal(swarms[i])
		}
	}

	for _, s := range swarms {
		s.AddSource(mult)
	}

	time.Sleep(10 * time.Second)
	log.Print("TEST: stopped sleeping")

	for _, s := range swarms {
		if s.BlockLen() < 2 {
			t.Fatal("Swarm BlockHistory is to short")
		}
	}
}
