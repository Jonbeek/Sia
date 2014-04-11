package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// Ensures that Tick() updates CurrentStep
func TestTick(t *testing.T) {
	// test takes 30 seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	s, err := CreateState(0)
	if err != nil {
		t.Fatal("Failed to create a state!")
	}

	// verify that tick is updating CurrentStep
	s.CurrentStep = 1
	go s.Tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	if s.CurrentStep == 1 {
		t.Fatal("s.CurrentStep failed to update correctly")
	}

	// verify that tick is wrapping around properly
	s.CurrentStep = common.QuorumSize
	time.Sleep(common.StepDuration)
	if s.CurrentStep != 1 {
		t.Fatal("s.CurrentStep failed to roll over")
	}

	// Plus one more test to make sure that a block-generate gets called
}

// An incomplete set of tests: the more complete suite will
// attack the system as a whole.
func TestHandleSignedHeartbeat(t *testing.T) {
	// create some public keys, 0 is self
	pubKey0, secKey0, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal("calling CreateKeyPair() failed!")
	}

	pubKey1, secKey1, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal("second call to CreateKeyPair() failed!")
	}

	// create SignedHeartbeat
	var sh SignedHeartbeat
	sh.Signatures = make([]crypto.Signature, 2)
	sh.Signatories = make([]ParticipantIndex, 2)

	// note: heartbeat never actually created

	// Create a set of signatures for the SignedHeartbeat
	signature0, err := crypto.Sign(secKey0, string(sh.HeartbeatHash[:]))
	if err != nil {
		t.Fatal("error signing HeartbeatHash")
	}

	signature1, err := crypto.Sign(secKey1, signature0.CombinedMessage())
	if err != nil {
		t.Fatal("error with second signing")
	}

	// build a valid SignedHeartbeat
	sh.Signatures[0] = signature0.Signature
	sh.Signatures[1] = signature1.Signature
	sh.Signatories[0] = 0
	sh.Signatories[1] = 1

	// create a state and populate it with the signatories as participants
	s, err := CreateState(0)
	if err != nil {
		t.Fatal("error creating state!")
	}

	s.AddParticipant(pubKey0, 0)
	s.AddParticipant(pubKey1, 1)

	// handle the signed heartbeat, expecting code 0
	returnCode := s.HandleSignedHeartbeat(&sh)
	if returnCode != 0 {
		t.Fatal("expected an invalid heartbeat:", returnCode)
	}

	// create a signed heartbeat with repeat signatures
	// create a heartbeat signed by a non-participant
	// send different heartbeats from same participant
	// send same second heartbeat multiple times...? (verify it doesn't get spammed out)
	// send heartbeats with invalid signatures
	// send heartbeats at invalid tick points
	// send a heartbeat right at the edge of a new block

	// all of this can be done without actually calling Tick()...

	// somehow verify that repeat heartbeats get ignored

	// somehow verify that new heartbeats get properly sent out
	// with valid signatures no less

	///////////////////

	// check that step timing if-else logic is correct
	// check that all signatures will verify
	// check that heartbeats are getting added to s.Heartbeats
}

// add fuzzing tests for HandleSignedHeartbeat
