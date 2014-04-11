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
	// create a state and populate it with the signatories as participants
	s, err := CreateState(0)
	if err != nil {
		t.Fatal("error creating state!")
	}

	// create some public keys
	pubKey1, secKey1, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal("calling CreateKeyPair() failed!")
	}

	pubKey2, secKey2, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal("second call to CreateKeyPair() failed!")
	}

	// Add keys as participants
	s.AddParticipant(pubKey1, 1)
	s.AddParticipant(pubKey2, 2)

	// create SignedHeartbeat
	var sh SignedHeartbeat
	sh.Heartbeat, err = s.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	sh.HeartbeatHash, err = crypto.CalculateTruncatedHash([]byte(sh.Heartbeat.Marshal()))
	if err != nil {
		t.Fatal(err)
	}
	sh.Signatures = make([]crypto.Signature, 2)
	sh.Signatories = make([]ParticipantIndex, 2)

	// Create a set of signatures for the SignedHeartbeat
	signature1, err := crypto.Sign(secKey1, string(sh.HeartbeatHash[:]))
	if err != nil {
		t.Fatal("error signing HeartbeatHash")
	}

	signature2, err := crypto.Sign(secKey2, signature1.CombinedMessage())
	if err != nil {
		t.Fatal("error with second signing")
	}

	// build a valid SignedHeartbeat
	sh.Signatures[0] = signature1.Signature
	sh.Signatures[1] = signature2.Signature
	sh.Signatories[0] = 1
	sh.Signatories[1] = 2

	// handle the signed heartbeat, expecting code 0
	returnCode := s.HandleSignedHeartbeat(&sh)
	if returnCode != 0 {
		t.Fatal("expected heartbeat to succeed:", returnCode)
	}

	// verify that a repeat heartbeat gets ignored
	returnCode = s.HandleSignedHeartbeat(&sh)
	if returnCode != 8 {
		t.Fatal("expected heartbeat to get ignored as a duplicate:", returnCode)
	}

	// create a signed heartbeat with repeat signatures
	// send heartbeats with invalid signatures
	// send heartbeats at invalid tick points
	// send a heartbeat right at the edge of a new block
	// somehow verify that new heartbeats get properly sent out with valid signatures
	// check that step timing if-else logic is correct
	// check that all signatures will verify
	// check that heartbeats are getting added to s.Heartbeats
}

// add fuzzing tests for HandleSignedHeartbeat
