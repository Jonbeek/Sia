package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// test create heartbeat

// test heartbeat.marshal and heartbeat.unmarshal

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

	// create a different heartbeat, this will be used to test the fail conditions
	sh.Heartbeat, err = s.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	sh.HeartbeatHash, err = crypto.CalculateTruncatedHash([]byte(sh.Heartbeat.Marshal()))
	if err != nil {
		t.Fatal(err)
	}

	// verify a heartbeat with bad signatures is rejected
	returnCode = s.HandleSignedHeartbeat(&sh)
	if returnCode != 6 {
		t.Fatal("expected heartbeat to get ignored as having invalid signatures: ", returnCode)
	}

	// give heartbeat repeat signatures
	signature1, err = crypto.Sign(secKey1, string(sh.HeartbeatHash[:]))
	if err != nil {
		t.Fatal("error with third signing")
	}

	signature2, err = crypto.Sign(secKey1, signature1.CombinedMessage())
	if err != nil {
		t.Fatal("error with fourth signing")
	}

	// adjust signatories slice
	sh.Signatures[0] = signature1.Signature
	sh.Signatures[1] = signature2.Signature
	sh.Signatories[0] = 1
	sh.Signatories[1] = 1

	// verify repeated signatures are rejected
	returnCode = s.HandleSignedHeartbeat(&sh)
	if returnCode != 5 {
		t.Fatal("expected heartbeat to be rejected for duplicate signatures: ", returnCode)
	}

	// remove second signature
	sh.Signatures = sh.Signatures[:1]
	sh.Signatories = sh.Signatories[:1]

	// handle heartbeat when tick is larger than num signatures
	s.CurrentStep = 2
	returnCode = s.HandleSignedHeartbeat(&sh)
	if returnCode != 2 {
		t.Fatal("expected heartbeat to be rejected as out-of-sync: ", returnCode)
	}

	// send a heartbeat right at the edge of a new block
	// test takes time; skip in short tests
	if testing.Short() {
		t.Skip()
	}

	// put block at edge
	s.CurrentStep = common.QuorumSize

	// submit heartbeat in separate thread
	go func() {
		returnCode = s.HandleSignedHeartbeat(&sh)
		if returnCode != 0 {
			t.Fatal("expected heartbeat to succeed!: ", returnCode)
		}
	}()

	time.Sleep(time.Second)
	s.CurrentStep = 1

	// verify that new heartbeats get properly sent out with valid signatures
}

// add fuzzing tests for HandleSignedHeartbeat
// test race conditions on HandleSignedHeartbeat

func TestProcessHeartbeat(t *testing.T) {
	// heh
}

func TestCompile(t *testing.T) {
	// confirm the ordering is consistent with the algorithm
	// this should probably just use a hard-coded example? not sure...
	//
	// check all participant stuffs in the for loop
	// have some empty participants, make sure 'continue' is reached
	// have participants without heartbeats
	// have participants with extra heartbeats
	// have participants  with 1 heartbeat, make sure process() gets called
	//
	// check that a new heartbeat is created
	// check that the SignedHeartbeat is correctly created
	//
	// check that heartbeat is properly sent to everybody who remains

	s, err := CreateState(0)
	if err != nil {
		t.Fatal(err)
	}
	s.Compile()
}

// Ensures that Tick() updates CurrentStep
func TestRegularTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	s, err := CreateState(0)
	if err != nil {
		t.Fatal(err)
	}

	// verify that tick is updating CurrentStep
	s.CurrentStep = 1
	go s.Tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	if s.CurrentStep != 2 {
		t.Fatal("s.CurrentStep failed to update correctly: ", s.CurrentStep)
	}
}

func TestCompilationTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	s, err := CreateState(0)
	if err != nil {
		t.Fatal(err)
	}

	// verify that tick is wrapping around properly
	s.CurrentStep = common.QuorumSize
	go s.Tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	if s.CurrentStep != 1 {
		t.Fatal("s.CurrentStep failed to roll over: ", s.CurrentStep)
	}

	// check if s.Compile() got called
}
