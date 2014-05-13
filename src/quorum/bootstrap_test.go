package quorum

import (
	"common"
	"testing"
)

// Bootstrap a state to the network, then another
func TestJoinQuorum(t *testing.T) {
	// Make a new state and network; start bootstrapping
	z := common.NewZeroNetwork()
	s0, err := CreateState(z)
	if err != nil {
		t.Fatal(err)
	}
	err = s0.JoinSia()
	if err != nil {
		t.Fatal(err)
	}

	// Verify the message for correctness

	// Forward message to bootstrap State (ourselves, as it were)
	m := z.RecentMessage(0)
	if m == nil {
		t.Fatal("message 0 never received")
	}
	s0.HandleJoinSia(m.Args.(Participant), nil)

	// Verify that a broadcast message went out indicating a new participant

	// Forward message to recipient
	m = z.RecentMessage(1)
	if m == nil {
		t.Fatal("message 1 never received")
	}
	s0.AddNewParticipant(m.Args.(Participant), nil)

	// Verify that we started ticking
	s0.tickingLock.Lock()
	if !s0.ticking {
		t.Fatal("Bootstrap state not ticking after joining Sia")
	}
	s0.tickingLock.Unlock()

	// Verify that s0.self.index updated
	if s0.self.index == 255 {
		t.Error("Bootstrapping failed to update State.self.index")
	}

	// Create a new state to bootstrap
	s1, err := CreateState(z)
	if err != nil {
		t.Fatal(err)
	}
	s1.JoinSia()

	// Verify message for correctness

	// Deliver message to bootstrap
	m = z.RecentMessage(2)
	s0.HandleJoinSia(m.Args.(Participant), nil)

	// Deliver the broadcasted messages
	m = z.RecentMessage(3)
	s0.AddNewParticipant(m.Args.(Participant), nil)
	m = z.RecentMessage(4)
	s1.AddNewParticipant(m.Args.(Participant), nil)

	// Verify the messages made it
	s1.tickingLock.Lock()
	if !s1.ticking {
		t.Error("s1 did not start ticking")
	}

	// both swarms should be aware of each other... maybe test their ongoing interactions?
}
