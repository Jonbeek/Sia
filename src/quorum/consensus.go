package quorum

import (
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
	Signatures []string
}

func (s *State) HandleSignedHeartbeat(sh *SignedHeartbeat) {
	// 1. look at the time and signature count, figure out if the hb is valid

	// 1. figure out if every signature is from a different host, but a host
	// 	that is represented in our state
	// 2. figure out if the update has arrived at an appropriate time -> 
	// 	all messages that arrive late are ignored, for security reasons.
}
