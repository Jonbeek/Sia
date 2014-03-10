package swarm

import (
	"common"
	"fmt"
	"log"
	"network"
	"testing"
	"time"
)

func BenchmarkHeatbeat(b *testing.B) {

	nodealive := make([]common.NetworkMessage, b.N)

	for i := 0; i < b.N; i++ {
		n, _ := common.RandomString(8)
		nodealive[i] = common.MarshalUpdate(&Heartbeat{Id: n})
	}

	bc := NewBlockchain("test", "f", time.Now().Add(1000*time.Second), make(map[string]interface{}))

	b.ResetTimer()

	for _, nm := range nodealive {
		bc.HandleNetworkMessage(nm)
	}

}

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

	time.Sleep(5 * time.Second)
	log.Print("TEST: stopped sleeping")

	informed := 0
	connected := 0

	for _, b := range swarms {
		switch s := b.GetState().(type) {
		case *StateSwarmInformed:
			s.Die()
			informed += 1
		case *StateSteady:
			s.Die()
			connected += 1
		case *ThreePhase:
			switch s.handler.(type) {
			case *StateSwarmInformed:
				informed += 1
			case *StateSteady:
				connected += 1
			default:
				t.Fatal(fmt.Sprintf("3phase? %T %T", b.GetState(), s.handler))
			}
		default:
			t.Fatal(fmt.Sprintf("Wrong type %T", b.GetState()))
		}
	}

	if connected <= common.SWARMSIZE/2 {
		t.Log("StateInformed", informed)
		t.Log("StateConnected", connected)
		t.Fatal("Failed to establish swarm")
	}

	for _, b := range swarms {
		if b.BlockLen() < 1 {
			t.Fatal("Blocks to short")
		}
	}
}
