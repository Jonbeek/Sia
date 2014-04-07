package quorum

import (
	"common"
	"time"
)

// This file will contain the code that manages concensus in the quorum
// The network needs to be guaranteed to be synched up. For now this will be done with timing.

// We need a ticker, that just moves over the variable for what step we are on.
// We can have two counters. The first is a counter of what step #, and the second
// is a counter of what step

// we're just waiting on messages at all times. When we get one, we'll call 'handle message'
// of some form

// blocks can only be recieved if it's time to get the block.

type SignedHeartbeat struct {
	HeartbeatHash string
	Signatures    []string
}

func (s *State) HandleSignedHeartbeat(sh *SignedHeartbeat) {
	// 1. figure out if every signature is from a different host, but a host
	// 	that is represented in our state
	// 2. figure out if the update has arrived at an appropriate time ->
	// 	all messages that arrive late are ignored, for security reasons.
	// 3. Validate all signatures, make sure they are signing a heartbeat with the correct parent block

	// need to worry about sync problems when we flip between blocks
}

// Tick() should only be called once, and should run in its own go thread
// Every common.SETPLENGTH, it updates the currentStep value.
// When the value flips from common.QUORUMSIZE to 1, Tick() calls
// 	integrateHeartbeats()
func (s *State) Tick() {
	// Every common.STEPLENGTH, advance the state stage
	ticker := time.Tick(common.STEPLENGTH)
	for _ = range ticker {
		if s.CurrentStep == common.QUORUMSIZE {
			// call logic to compile block
			s.CurrentStep = 1
		} else {
			s.CurrentStep += 1
		}
	}
}
