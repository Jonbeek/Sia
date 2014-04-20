package quorum

import (
	"common"
	"common/crypto"
	"testing"
)

// quick sanity check
func TestCreateState(t *testing.T) {
	// create a state
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	// verify that the keys can sign and be verified
	err = crypto.CheckKeys(s.participants[s.participantIndex].PublicKey, s.secretKey)
	if err != nil {
		t.Fatal(err)
	}

	// sanity check CurrentStep
	if s.currentStep != 1 {
		t.Fatal("Current step should be initialized to 1!")
	}
}

// verify that one state can add another
func TestAddParticipant(t *testing.T) {
	s0, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	s1, err := CreateState(common.NewZeroNetwork(), 1)
	if err != nil {
		t.Fatal(err)
	}

	err = s0.AddParticipant(s1.Self(), 1)
	if err != nil {
		t.Fatal(err)
	}

	// check that participant 1 was added to state 0
	if s1.participants[s1.participantIndex].PublicKey != s0.participants[1].PublicKey {
		t.Fatal("AddParticipant failed!")
	}
}

// check general case, check corner cases, and then do some fuzzing
func TestrandInt(t *testing.T) {
	s, err := CreateState(common.NewZeroNetwork(), 0)
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
	for i := 0; i < 10000; i++ {
		randInt, err = s.randInt(low, high)
		if err != nil {
			t.Fatal("randInt fuzzing error: ", err)
		}

		if randInt < low || randInt >= high {
			t.Fatal("randInt fuzzing: ", randInt, " produced, expected number between ", low, " and ", high)
		}
	}
}
