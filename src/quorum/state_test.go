package quorum

/*
import (
	"common"
	"common/crypto"
	"testing"
)

// Verify zero case marshalling works, check inputs, do fuzzing
 func TestParticipantMarshalling(t *testing.T) {
	// zero case marshalling
	p := new(participant)
	mp := p.marshal()
	up, err := unmarshalParticipant(mp)
	if err != nil {
		t.Fatal(err)
	}
	if *up != *p {
		t.Fatal("Zero case marshalling and unmarshalling not equal")
	}

	// Attempt bad input
	var bad []byte
	up, err = unmarshalParticipant(bad)
	if err == nil {
		t.Fatal("unmarshalled an empty []byte")
	}
	bad = make([]byte, crypto.PublicKeySize+4)
	up, err = unmarshalParticipant(bad)
	if err == nil {
		t.Fatal("unmarshalled a []byte of insufficient length")
	}

	// fuzzing
}

// Create a state, check the defaults
func TestCreateState(t *testing.T) {
	// does a state create without errors?
	s, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// check that previousEntropyStage1 is initialized correctly
	var emptyEntropy common.Entropy
	emptyHash, err := crypto.CalculateTruncatedHash(emptyEntropy[:])
	if err != nil {
		t.Fatal(err)
	}
	for i := range s.previousEntropyStage1 {
		if s.previousEntropyStage1[i] != emptyHash {
			t.Error("previousEntropyStage1 initialized incorrectly at index ", i)
		}
	}

	// sanity check the default values
	if s.participantIndex != 255 {
		t.Error("s.participantIndex initialized to ", s.participantIndex)
	}
	if s.currentStep != 1 {
		t.Error("s.currentStep should be initialized to 1!")
	}
	if s.wallets == nil {
		t.Error("s.wallets was not initialized")
	}
}

// Bootstrap a state to the network, then another
func TestJoinQuorum(t *testing.T) {
	// Make a new state and network; start bootstrapping
	z := common.NewZeroNetwork()
	s0, err := CreateState(z)
	if err != nil {
		t.Fatal(err)
	}
	s0.JoinSia()

	// Verify the message for correctness

	// Forward message to bootstrap State (ourselves, as it were)
	s0.HandleMessage(z.RecentMessage(0).Payload)

	// Verify that a broadcast message went out indicating a new participant

	// Forward message to recipient
	s0.HandleMessage(z.RecentMessage(1).Payload)

	// Verify that we started ticking
	s0.tickingLock.Lock()
	if !s0.ticking {
		t.Error("Bootstrap state not ticking after joining Sia")
	}
	s0.tickingLock.Unlock()

	// Verify that s0.participantIndex updated
	if s0.participantIndex == 255 {
		t.Error("Bootstrapping failed to update State.participantIndex")
	}

	// Create a new state to bootstrap
	s1, err := CreateState(z)
	if err != nil {
		t.Fatal(err)
	}
	s1.JoinSia()

	// Verify message for correctness

	// Deliver message to bootstrap
	s0.HandleMessage(z.RecentMessage(2).Payload)

	// Deliver the broadcasted messages
	s0.HandleMessage(z.RecentMessage(3).Payload)
	s1.HandleMessage(z.RecentMessage(4).Payload)

	// Verify the messages made it
	s1.tickingLock.Lock()
	if !s1.ticking {
		t.Error("s1 did not start ticking")
	}

	// both swarms should be aware of each other... maybe test their ongoing interactions?
}

// test HandleMessage and SetAddress

// check general case, check corner cases, and then do some fuzzing
func TestRandInt(t *testing.T) {
	s, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// check that it works in the vanilla case
	previousEntropy := s.currentEntropy
	randInt, err := s.randInt(0, 5)
	if err != nil {
		t.Fatal(err)
	}
	if randInt < 0 || randInt >= 5 {
		t.Fatal("randInt returned but is not between the bounds")
	}

	// check that s.CurrentEntropy flipped to next value
	if previousEntropy == s.currentEntropy {
		t.Fatal("When calling randInt, s.CurrentEntropy was not changed")
	}

	// check the zero value
	randInt, err = s.randInt(0, 0)
	if err == nil {
		t.Fatal("Randint(0,0) should return a bounds error")
	}

	// fuzzing, skip for short tests
	if testing.Short() {
		t.Skip()
	}

	low := 0
	high := common.QuorumSize
	for i := 0; i < 100000; i++ {
		randInt, err = s.randInt(low, high)
		if err != nil {
			t.Fatal("randInt fuzzing error: ", err)
		}

		if randInt < low || randInt >= high {
			t.Fatal("randInt fuzzing: ", randInt, " produced, expected number between ", low, " and ", high)
		}
	}
} */
