package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// Verify that newHeartbeat() produces valid heartbeats
func TestNewHeartbeat(t *testing.T) {
	// create a state, and then a heartbeat
	s, err := CreateState(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	hb, err := s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}

	// verify that entropy is being properly generated when making the heartbeat
	storedEntropyHash, err := crypto.CalculateTruncatedHash(s.storedEntropyStage2[:])
	if err != nil {
		t.Fatal(err)
	} else if hb.entropyStage1 != storedEntropyHash {
		t.Fatal("newHeartbeat() incorrectly producing EntropyStage1 from s.StoredEntropyStage2")
	}

	// verify that hosts accept the new heartbeats
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
	if hb.entropyStage1 != uhb.entropyStage1 {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}
	if hb.entropyStage2 != uhb.entropyStage2 {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}

	// test encoding with bad input
	hb = nil
	mhb, err = hb.GobEncode()
	if err == nil {
		t.Error("able to encode a nil heartbeat")
	}
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
		t.Fatal("error signing HeartbeatHash")
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
		t.Fatal("expected heartbeat to get ignored as a duplicate:", err)
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

	/*// give heartbeat repeat signatures
	signature1, err = crypto.Sign(secKey1, string(sh.heartbeatHash[:]))
	if err != nil {
		t.Fatal("error with third signing")
	}

	signature2, err = crypto.Sign(secKey1, signature1.CombinedMessage())
	if err != nil {
		t.Fatal("error with fourth signing")
	}

	// adjust signatories slice
	sh.signatures[0] = signature1.Signature
	sh.signatures[1] = signature2.Signature
	sh.signatories[0] = 1
	sh.signatories[1] = 1

	// verify repeated signatures are rejected
	msh, err = sh.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.handleSignedHeartbeat(msh)
	if returnCode != 5 {
		t.Fatal("expected heartbeat to be rejected for duplicate signatures: ", returnCode)
	}

	// remove second signature
	sh.signatures = sh.signatures[:1]
	sh.signatories = sh.signatories[:1]

	// handle heartbeat when tick is larger than num signatures
	s.currentStep = 2
	msh, err = sh.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.handleSignedHeartbeat(msh)
	if returnCode != 2 {
		t.Fatal("expected heartbeat to be rejected as out-of-sync: ", returnCode)
	}

	// send a heartbeat right at the edge of a new block
	// test takes time; skip in short tests
	if testing.Short() {
		t.Skip()
	}

	// put block at edge
	s.currentStep = common.QuorumSize

	// submit heartbeat in separate thread
	go func() {
		msh, err = sh.marshal()
		if err != nil {
			t.Fatal(err)
		}
		returnCode = s.handleSignedHeartbeat(msh)
		if returnCode != 0 {
			t.Fatal("expected heartbeat to succeed!: ", returnCode)
		}
	}()

	time.Sleep(time.Second)
	*/
}

// add fuzzing tests for HandleSignedHeartbeat
// test race conditions on HandleSignedHeartbeat

/*func TestTossParticipant(t *testing.T) {
	// tossParticipant isn't yet implemented
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
	s0.AddParticipant(s1.Self(), 1)
	s1.AddParticipant(s0.Self(), 0)

	// check that a valid heartbeat passes
	hb0, err := s0.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	returnCode := s1.processHeartbeat(hb0, 0)
	if returnCode != 0 {
		t.Fatal("processHeartbeat threw out a valid heartbeat")
	}

	// check that invalid entropy fails
	hb1, err := s1.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	hb1.entropyStage2[0] = 1
	returnCode = s0.processHeartbeat(hb1, 1)
	if returnCode != 1 {
		t.Fatal("processHeartbeat accepted an invalid heartbeat")
	}
}*/

// TestCompile should probably be reviewed and rehashed
/* func TestCompile(t *testing.T) {
	// Create states and add them to eachother as participants
	s0, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}
	s1, err := CreateState(common.NewZeroNetwork(), 1)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := CreateState(common.NewZeroNetwork(), 2)
	if err != nil {
		t.Fatal(err)
	}
	s0.AddParticipant(s1.Self(), 1)
	s0.AddParticipant(s2.Self(), 2)

	// fetch legal heartbeat for s0
	hb0, err := s0.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	shb0, err := s0.signHeartbeat(hb0)
	if err != nil {
		t.Fatal(err)
	}

	// fetch legal heartbeat for s2
	hb2a, err := s2.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	shb2a, err := s2.signHeartbeat(hb2a)
	if err != nil {
		t.Fatal(err)
	}

	// create a second illegal heartbeat for s2
	var hb2b heartbeat
	hb2b.entropyStage2 = hb2a.entropyStage2
	shb2b, err := s2.signHeartbeat(&hb2b)
	if err != nil {
		t.Fatal(err)
	}

	// send the SignedHeartbeats to s0
	mshb0, err := shb0.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode := s0.handleSignedHeartbeat(mshb0)
	if returnCode != 0 {
		t.Fatal("Expecting shb0 to be valid: ", returnCode)
	}
	mshb2a, err := shb2a.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s0.handleSignedHeartbeat(mshb2a)
	if returnCode != 0 {
		t.Fatal("Expecting shb2a to be valid: ", returnCode)
	}
	mshb2b, err := shb2b.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s0.handleSignedHeartbeat(mshb2b)
	if returnCode != 0 {
		t.Fatal("Expecting shb2b to be valid: ", returnCode)
	}

	s0.compile()

	// check that hosts arrive at the same participantOrdering
	participantOrdering1 := s1.participantOrdering()
	participantOrdering2 := s2.participantOrdering()
	if participantOrdering1 != participantOrdering2 {
		t.Fatal("partcipantOrderings for s1 and s2 are not identical!")
	}

	// verify that upon processing, s0 is not thrown from s0, and is processed correctly
	if s0.participants[0] == nil {
		t.Fatal("s0 thrown from s0 despite having a fair heartbeat")
	}

	// verify that upon processing, s1 is thrown from s0 (doesn't have heartbeat)
	if s0.participants[1] != nil {
		t.Fatal("s1 not thrown from s0 despite having no heartbeats")
	}

	// verify that upon processing, s3 is thrown from s0 (too many heartbeats)
	if s0.participants[2] != nil {
		t.Fatal("s2 not thrown from s0 despite having multiple heartbeats")
	}

	// verify that a new heartbeat was made, formatted into a SignedHeartbeat, and sent off
} */

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
