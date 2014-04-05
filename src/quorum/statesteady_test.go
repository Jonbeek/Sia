package quorum

import (
	"common"
	"log"
	"testing"
	"time"
)

func TestStateSteady(t *testing.T) {

	old := common.SWARMSIZE
	common.SWARMSIZE = 4
	defer func(old int) {
		common.SWARMSIZE = old
	}(old)

	swarms, _ := CreateSteadySwarm(t)

	log.SetFlags(log.Lmicroseconds)

	time.Sleep(5 * time.Second)
	log.Print("TEST: stopped sleeping")

	for _, s := range swarms {
		s.GetState().Die()
	}

	for _, s := range swarms {
		if s.BlockLen() < 2 {
			t.Fatal("Swarm BlockHistory is to short")
		}
	}
}
