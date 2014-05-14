package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// Verify that newHeartbeat() produces valid heartbeats
func TestNewHeartbeat(t *testing.T) {
	// tbi
}

func TestHeartbeatEncoding(t *testing.T) {
	// marshal an empty heartbeat
	hb := new(heartbeat)
	mhb, err := hb.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// unmarshal the empty heartbeat
	uhb := new(heartbeat)
	err = uhb.GobDecode(mhb)
	if err != nil {
		t.Fatal(err)
	}

	// test for equivalency
	if hb.entropy != uhb.entropy {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}

	// test encoding with bad input
	err = uhb.GobDecode(nil)
	if err == nil {
		t.Error("able to decode a nil byte slice")
	}

	// fuzz over random potential values of heartbeat
}

func TestSignHeartbeat(t *testing.T) {
	// tbi
}

func TestSignedHeartbeatEncoding(t *testing.T) {
	// Test for bad inputs
	var bad *SignedHeartbeat
	bad = nil
	_, err := bad.GobEncode()
	if err == nil {
		t.Error("Should not encode a nil signedHeartbeat")
	}
	err = bad.GobDecode(nil)
	if err == nil {
		t.Error("Should not be able to decode a nil byte slice")
	}

	// Test the encoding and decoding of a simple signed heartbeat
	s, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	hb, err := s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	sh, err := s.signHeartbeat(hb)
	if err != nil {
		t.Fatal(err)
	}
	msh, err := sh.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	ush := new(SignedHeartbeat)
	err = ush.GobDecode(msh)
	if err != nil {
		t.Fatal(err)
	}

	// check encoding and decoding of a signedHeartbeat with many signatures
}

func TestHandleSignedHeartbeat(t *testing.T) {
	// create a state and populate it with the signatories as participants
	s, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// create keypairs
	pubKey1, secKey1, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, secKey2, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// create participants and add them to s
	var p1 Participant
	var p2 Participant
	p1.index = 1
	p2.index = 2
	p1.publicKey = pubKey1
	p2.publicKey = pubKey2
	err = s.AddNewParticipant(p1, nil)
	if err != nil {
		t.Fatal(err)
	}
	s.AddNewParticipant(p2, nil)
	if err != nil {
		t.Fatal(err)
	}

	// create SignedHeartbeat
	var sh SignedHeartbeat
	sh.heartbeat, err = s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	esh, err := sh.heartbeat.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash(esh)
	if err != nil {
		t.Fatal(err)
	}
	sh.signatures = make([]crypto.Signature, 2)
	sh.signatories = make([]byte, 2)

	// Create a set of signatures for the SignedHeartbeat
	signature1, err := secKey1.Sign(sh.heartbeatHash[:])
	if err != nil {
		t.Fatal(err)
	}

	combinedMessage, err := signature1.CombinedMessage()
	if err != nil {
		t.Fatal(err)
	}
	signature2, err := secKey2.Sign(combinedMessage)
	if err != nil {
		t.Fatal(err)
	}

	// build a valid SignedHeartbeat
	sh.signatures[0] = signature1.Signature
	sh.signatures[1] = signature2.Signature
	sh.signatories[0] = 1
	sh.signatories[1] = 2

	// delete existing heartbeat from state; makes the remaining tests easier
	s.heartbeats[sh.signatories[0]] = make(map[crypto.TruncatedHash]*heartbeat)

	// handle the signed heartbeat, expecting nil error
	err = s.HandleSignedHeartbeat(sh, nil)
	if err != nil {
		t.Fatal(err)
	}

	// verify that a repeat heartbeat gets ignored
	err = s.HandleSignedHeartbeat(sh, nil)
	if err != hsherrHaveHeartbeat {
		t.Error("expected heartbeat to get ignored as a duplicate:", err)
	}

	// create a different heartbeat, this will be used to test the fail conditions
	sh.heartbeat, err = s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	ehb, err := sh.heartbeat.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash(ehb)
	if err != nil {
		t.Fatal(err)
	}

	// verify a heartbeat with bad signatures is rejected
	err = s.HandleSignedHeartbeat(sh, nil)
	if err != hsherrInvalidSignature {
		t.Error("expected heartbeat to get ignored as having invalid signatures: ", err)
	}

	// verify that a non-participant gets rejected
	sh.signatories[0] = 3
	err = s.HandleSignedHeartbeat(sh, nil)
	if err != hsherrNonParticipant {
		t.Error("expected non-participant to be rejected: ", err)
	}

	// give heartbeat repeat signatures
	signature1, err = secKey1.Sign(sh.heartbeatHash[:])
	if err != nil {
		t.Fatal(err)
	}

	combinedMessage, err = signature1.CombinedMessage()
	if err != nil {
		t.Fatal(err)
	}
	signature2, err = secKey1.Sign(combinedMessage)
	if err != nil {
		t.Error(err)
	}

	// adjust signatories slice
	sh.signatures[0] = signature1.Signature
	sh.signatures[1] = signature2.Signature
	sh.signatories[0] = 1
	sh.signatories[1] = 1

	// verify repeated signatures are rejected
	err = s.HandleSignedHeartbeat(sh, nil)
	if err != hsherrDoubleSigned {
		t.Error("expected heartbeat to be rejected for duplicate signatures: ", err)
	}

	// remove second signature
	sh.signatures = sh.signatures[:1]
	sh.signatories = sh.signatories[:1]

	// handle heartbeat when tick is larger than num signatures
	s.stepLock.Lock()
	s.currentStep = 2
	s.stepLock.Unlock()
	err = s.HandleSignedHeartbeat(sh, nil)
	if err != hsherrNoSync {
		t.Error("expected heartbeat to be rejected as out-of-sync: ", err)
	}

	// remaining tests require sleep
	if testing.Short() {
		t.Skip()
	}

	// send a heartbeat right at the edge of a new block
	s.stepLock.Lock()
	s.currentStep = common.QuorumSize
	s.stepLock.Unlock()

	// submit heartbeat in separate thread
	go func() {
		err = s.HandleSignedHeartbeat(sh, nil)
		if err != nil {
			t.Fatal("expected heartbeat to succeed!: ", err)
		}
		// need some way to verify with the test that the funcion gets here
	}()

	s.stepLock.Lock()
	s.currentStep = 1
	s.stepLock.Unlock()
	time.Sleep(time.Second)
	time.Sleep(common.StepDuration)
}

func TestTossParticipant(t *testing.T) {
	// tbi
}

// Check that valid heartbeats are accepted and invalid heartbeats are rejected
func TestProcessHeartbeat(t *testing.T) {
	// create states and add them to each other
	s0, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	s1, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	s0.AddNewParticipant(*s1.self, nil)
	s1.AddNewParticipant(*s0.self, nil)

	// check that a valid heartbeat passes
	hb0, err := s0.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	err = s1.processHeartbeat(hb0, 0)
	if err != nil {
		t.Error("processHeartbeat threw out a valid heartbeat: ", err)
	}
}

// TestCompile should probably be reviewed and rehashed
func TestCompile(t *testing.T) {
	// tbi
}

// Ensures that Tick() updates CurrentStep
func TestRegularTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	s, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// verify that tick is updating CurrentStep
	s.stepLock.Lock()
	s.currentStep = 1
	s.stepLock.Unlock()
	go s.tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	s.stepLock.Lock()
	if s.currentStep != 2 {
		t.Fatal("s.currentStep failed to update correctly: ", s.currentStep)
	}
	s.stepLock.Unlock()
}

// ensures Tick() calles compile() and then resets the counter to step 1
func TestCompilationTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	// create state, set values for compile
	s, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	s.currentStep = common.QuorumSize
	go s.tick()

	// verify that tick is wrapping around properly
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	s.stepLock.Lock()
	if s.currentStep != 1 {
		t.Error("s.currentStep failed to roll over: ", s.currentStep)
	}
	s.stepLock.Unlock()
}
