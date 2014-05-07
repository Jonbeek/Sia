package main

import (
	"common"
	"network"
	"quorum"
	"testing"
	"time"
)

func TestNetworkedQuorum(t *testing.T) {
	// create a MessageRouter and 4 participants
	rpcs, err := network.NewRPCServer(9988)
	if err != nil {
		println("message sender creation failed")
	}

	s0, err := quorum.CreateState(rpcs)
	if err != nil {
		println("s0 creation failed")
	}
	s1, err := quorum.CreateState(rpcs)
	if err != nil {
		println("s1 creation failed")
	}
	s2, err := quorum.CreateState(rpcs)
	if err != nil {
		println("s2 creation failed")
	}
	s3, err := quorum.CreateState(rpcs)
	if err != nil {
		println("s3 creation failed")
	}

	s0.JoinSia()
	s1.JoinSia()
	s2.JoinSia()
	s3.JoinSia()

	// Basically checking for errors up to this point
	if testing.Short() {
		t.Skip()
	}

	time.Sleep(3 * common.StepDuration * time.Duration(common.QuorumSize))

	// if no seg faults, no errors
	// there needs to be a s0.ParticipantStatus() call returning a function with public information about the participant
	// there needs to be a s0.QuorumStatus() call returning public information about the quorum
	// 		all participants in a public quorum should return the same information
}
