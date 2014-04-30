package quorum

import (
	"common"
	"common/crypto"
	"testing"
)

func TestParticipantCompare(t *testing.T) {
	var p0 *participant
	var p1 *participant
	p0 = nil
	p1 = nil

	// compare nil values
	compare := p0.compare(p1)
	if compare == true {
		t.Error("Comparing any nil participant should return false")
	}

	// compare when one is nil
	p0 = new(participant)
	compare = p0.compare(p1)
	if compare == true {
		t.Error("Comparing a zero participant to a nil participant should return false")
	}
	compare = p1.compare(p0)
	if compare == true {
		t.Error("Comparing a zero participant to a nil participant should return false")
	}

	// initialize each participant with a public key
	p1 = new(participant)
	pubKey, _, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p0.publicKey = pubKey
	p1.publicKey = pubKey

	// compare initialized participants
	compare = p0.compare(p1)
	if compare == false {
		t.Error("Comparing two zero participants should return true")
	}
	compare = p1.compare(p0)
	if compare == false {
		t.Error("Comparing two zero participants should return true")
	}

	// compare when address are not equal (piecewise for each component of address

	// compare when public keys are not equivalent
	pubKey, _, err = crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p1.publicKey = pubKey
	compare = p0.compare(p1)
	if compare == true {
		t.Error("Comparing two participants with different public keys should return false")
	}
	compare = p1.compare(p0)
	if compare == true {
		t.Error("Comparing two participants with different public keys should return false")
	}
}

func TestParticipantMarshalling(t *testing.T) {
	// zero case marshalling
	p := new(participant)
	up := new(participant)
	_, err := p.GobEncode()
	if err == nil {
		t.Fatal("Should not be able to encode nil values")
	}

	// bad input for marshalling and unmarshalling

	// Make a bootstrap participant
	pubKey, _, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p.publicKey = pubKey
	p.address.Id = bootstrapId
	p.address.Host = bootstrapHost
	p.address.Port = bootstrapPort

	mp, err := p.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	err = up.GobDecode(mp)
	if err != nil {
		t.Fatal(err)
	}

	if up.address != p.address {
		t.Error("up.address != p.address")
	}

	compare := up.publicKey.Compare(&p.publicKey)
	if compare != true {
		t.Error("up.PublicKey != p.PublicKey")
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
	s0.HandleMessage(m.Payload)

	// Verify that a broadcast message went out indicating a new participant

	// Forward message to recipient
	m = z.RecentMessage(1)
	if m == nil {
		t.Fatal("message 1 never received")
	}
	s0.HandleMessage(m.Payload)

	// Verify that we started ticking
	s0.tickingLock.Lock()
	if !s0.ticking {
		t.Fatal("Bootstrap state not ticking after joining Sia")
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

func TestSetAddress(t *testing.T) {
	// ?
}

func TestUpdateParticipant(t *testing.T) {
	// ?
}

func TestHandleMessage(t *testing.T) {
	// ?
}

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
		t.Error(previousEntropy)
		t.Error(s.currentEntropy)
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
			t.Fatal("randInt fuzzing error: ", err, " low: ", low, " high: ", high)
		}

		if randInt < low || randInt >= high {
			t.Fatal("randInt fuzzing: ", randInt, " produced, expected number between ", low, " and ", high)
		}
	}
}
