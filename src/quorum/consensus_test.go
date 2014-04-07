package quorum

import (
	"common"
	"testing"
	"time"
)

// Only to be used in long tests
// Ensures that Tick() updates CurrentStep
func TestTick(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var s State
	s.CurrentStep = 1
	go s.Tick()
	time.Sleep(common.STEPLENGTH)
	time.Sleep(time.Second)

	if s.CurrentStep == 1 {
		t.Fatal("s.CurrentStep failed to update correctly!")
	}

	s.CurrentStep = common.QUORUMSIZE
	time.Sleep(common.STEPLENGTH)
	if s.CurrentStep != 1 {
		t.Fatal("s.CurrentStep failed to roll over!")
	}

	// Plus one more test to make sure that a block-generate gets called
}

// Testing for HandleSignedHeartbeat

// Testing for NewHeartbeat
