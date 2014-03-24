package swarm

import (
	"common"
	"log"
	"testing"
	"time"
)

func TestStateJoin(t *testing.T) {

	old := common.SWARMSIZE
	common.SWARMSIZE = 4
	defer func(old int) {
		common.SWARMSIZE = old
	}(old)

	log.SetFlags(log.Lmicroseconds)

	swarms, mult := CreateSteadySwarm(t)

	log.Print("TEST: stopped sleeping")

	host := "joininghost"

	b := JoinBlockchain(host, swarms[0].Id, swarms[0].LastBlock().Time.Add(4*time.Second), nil)
	b.AddSource(mult)

	time.Sleep(9 * time.Second)

	for _, s := range swarms {
		s.GetState().Die()
	}

	if !swarms[0].HostActive(host) {
		log.Fatal("Host failed to join")
	}

}
