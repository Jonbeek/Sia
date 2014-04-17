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
	err = crypto.CheckKeys(s.Participants[s.ParticipantIndex].PublicKey, s.SecretKey)
	if err != nil {
		t.Fatal(err)
	}

	// sanity check CurrentStep
	if s.CurrentStep != 1 {
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
	if s1.Participants[s1.ParticipantIndex].PublicKey != s0.Participants[1].PublicKey {
		t.Fatal("AddParticipant failed!")
	}
}

// check general case, check corner cases, and then do some fuzzing
func TestRandInt(t *testing.T) {
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	// check that it works in the vanilla case
	previousEntropy := s.CurrentEntropy
	randInt, err := s.RandInt(0, 5)
	if err != nil {
		t.Fatal(err)
	}
	if randInt < 0 || randInt >= 5 {
		t.Fatal("randInt returned but is not between the bounds")
	}

	// check that s.CurrentEntropy flipped to next value
	if previousEntropy == s.CurrentEntropy {
		t.Fatal("When calling RandInt, s.CurrentEntropy was not changed")
	}

	// check the zero value
	randInt, err = s.RandInt(0, 0)
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
		randInt, err = s.RandInt(low, high)
		if err != nil {
			t.Fatal("RandInt fuzzing error: ", err)
		}

		if randInt < low || randInt >= high {
			t.Fatal("RandInt fuzzing: ", randInt, " produced, expected number between ", low, " and ", high)
		}
	}
}
