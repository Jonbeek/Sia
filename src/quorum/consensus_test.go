package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// Verify that newHeartbeat() produces valid heartbeats
func TestnewHeartbeat(t *testing.T) {
	// create a state, and then a heartbeat
	s, err := CreateState(common.NewZeroNetwork(), 0)
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
}

// Marshalling and Unmarshalling should result in equivalent Heartbeats
func TestHeartbeatMarshalling(t *testing.T) {
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	// verify that a heartbeat, once unmarshalled, is identical to the original
	hbOriginal, err := s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	hbMarshalled := hbOriginal.marshal()
	hbUnmarshalled, err := unmarshalHeartbeat(hbMarshalled)
	if err != nil {
		t.Fatal(err)
	}
	if hbOriginal.entropyStage1 != hbUnmarshalled.entropyStage1 {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}
	if hbOriginal.entropyStage2 != hbUnmarshalled.entropyStage2 {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}

	// verify that input is being checked for UnmarshalHeartbeat
	_, err = unmarshalHeartbeat(hbMarshalled[1:])
	if err == nil {
		t.Fatal("Heartbeat unmarshalling succeded with a short input")
	}
	_, err = unmarshalHeartbeat(append(hbMarshalled, hbMarshalled...))
	if err == nil {
		t.Fatal("Heartbeat unmarshalling succeded with a long input")
	}
}

// a SignedHeartbeat should be the same after marshalling and unmarshalling
func TestSignedHeartbeatMarshalling(t *testing.T) {
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	// create a SignedHeartbeat, marshall and unmarshall it, and test equivalency
	hb, err := s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	originalSignedHeartbeat, err := s.signHeartbeat(hb)
	if err != nil {
		t.Fatal(err)
	}
	marshalledSH, err := originalSignedHeartbeat.marshal()
	if err != nil {
		t.Fatal(err)
	}
	unmarshalledSH, err := unmarshalSignedHeartbeat(marshalledSH)
	if string(unmarshalledSH.heartbeat.marshal()) != string(originalSignedHeartbeat.heartbeat.marshal()) {
		t.Fatal("Heartbeat changed after being marshalled and unmarshalled")
	}
	if unmarshalledSH.heartbeatHash != originalSignedHeartbeat.heartbeatHash {
		t.Fatal("HeartbeatHash changed after being marshalled and unmarshalled")
	}
	if len(unmarshalledSH.signatures) != len(originalSignedHeartbeat.signatures) {
		t.Fatal("Length of SignedHeartbeat.Signatures changed after being marshalled and unmarshalled")
	}
	if len(unmarshalledSH.signatories) != len(originalSignedHeartbeat.signatories) {
		t.Fatal("Length of SignedHeartbeat.Signatories changed after being marshalled and unmarshalled")
	}
	for i := 0; i < len(unmarshalledSH.signatures); i++ {
		if unmarshalledSH.signatures[i] != originalSignedHeartbeat.signatures[i] {
			t.Fatal("For i=", i, ", unmarshalledSH.Signatures[i] did not equal originalSignedHeartbeat.Signatures[i]")
		}
		if unmarshalledSH.signatories[i] != originalSignedHeartbeat.signatories[i] {
			t.Fatal("For i=", i, ", unmarshalledSH.Signatories[i] did not equal originalSignedHeartbeat.Signatories[i]")
		}
	}

	// verify that input is being checked for SignedHeartbeat.Marshal()
	originalSignedHeartbeat.signatures = make([]crypto.Signature, 2*common.QuorumSize)
	_, err = originalSignedHeartbeat.marshal()
	if err == nil {
		t.Fatal("SignedHeartbeat.Marshal() needs to check the length of msh.Signatures")
	}
	originalSignedHeartbeat.signatures = make([]crypto.Signature, 2)
	if err == nil {
		t.Fatal("SignedHeartbeat.Marshal() needs to check that msh.Signatures and msh.Signatories are equal in length")
	}

	// verify that input is being checked for UnmarshalSignedHeartbeat()
	_, err = unmarshalSignedHeartbeat(marshalledSH[:marshalledHeartbeatLen()])
	if err == nil {
		t.Fatal("UnmarshalSignedHeartbeat not checking input length")
	}
	_, err = unmarshalSignedHeartbeat(marshalledSH[1:])
	if err == nil {
		t.Fatal("UnmarshalSignedHeartbeat succeeded when input was too short")
	}
	_, err = unmarshalSignedHeartbeat(append(marshalledSH, marshalledSH...))
	if err == nil {
		t.Fatal("UnmarshalSignedHeartbeat succeded when input was too long")
	}
}

// TestHandleSignedHeartbeat should probably be reviewed and rehashed
func TestHandleSignedHeartbeat(t *testing.T) {
	// create a state and populate it with the signatories as participants
	s, err := CreateState(common.NewZeroNetwork(), 0)
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
	p1 := new(Participant)
	p2 := new(Participant)
	p1.PublicKey = pubKey1
	p2.PublicKey = pubKey2
	s.AddParticipant(p1, 1)
	s.AddParticipant(p2, 2)

	// create SignedHeartbeat
	var sh signedHeartbeat
	sh.heartbeat, err = s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash([]byte(sh.heartbeat.marshal()))
	if err != nil {
		t.Fatal(err)
	}
	sh.signatures = make([]crypto.Signature, 2)
	sh.signatories = make([]participantIndex, 2)

	// Create a set of signatures for the SignedHeartbeat
	signature1, err := crypto.Sign(secKey1, string(sh.heartbeatHash[:]))
	if err != nil {
		t.Fatal("error signing HeartbeatHash")
	}

	signature2, err := crypto.Sign(secKey2, signature1.CombinedMessage())
	if err != nil {
		t.Fatal("error with second signing")
	}

	// build a valid SignedHeartbeat
	sh.signatures[0] = signature1.Signature
	sh.signatures[1] = signature2.Signature
	sh.signatories[0] = 1
	sh.signatories[1] = 2

	// handle the signed heartbeat, expecting code 0
	msh, err := sh.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode := s.handleSignedHeartbeat(msh)
	if returnCode != 0 {
		t.Fatal("expected heartbeat to succeed:", returnCode)
	}

	// verify that a repeat heartbeat gets ignored
	msh, err = sh.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.handleSignedHeartbeat(msh)
	if returnCode != 8 {
		t.Fatal("expected heartbeat to get ignored as a duplicate:", returnCode)
	}

	// create a different heartbeat, this will be used to test the fail conditions
	sh.heartbeat, err = s.newHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash([]byte(sh.heartbeat.marshal()))
	if err != nil {
		t.Fatal(err)
	}

	// verify a heartbeat with bad signatures is rejected
	msh, err = sh.marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.handleSignedHeartbeat(msh)
	if returnCode != 6 {
		t.Fatal("expected heartbeat to get ignored as having invalid signatures: ", returnCode)
	}

	// give heartbeat repeat signatures
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
}

// add fuzzing tests for HandleSignedHeartbeat
// test race conditions on HandleSignedHeartbeat

func TestTossParticipant(t *testing.T) {
	// tossParticipant isn't yet implemented
}

// Check that valid heartbeats are accepted and invalid heartbeats are rejected
func TestProcessHeartbeat(t *testing.T) {
	// create states and add them to each other
	s0, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}
	s1, err := CreateState(common.NewZeroNetwork(), 1)
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
}

// TestCompile should probably be reviewed and rehashed
func TestCompile(t *testing.T) {
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
}

// Ensures that Tick() updates CurrentStep
func TestRegularTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	// verify that tick is updating CurrentStep
	s.currentStep = 1
	go s.tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	s.lock.Lock()
	if s.currentStep != 2 {
		t.Fatal("s.currentStep failed to update correctly: ", s.currentStep)
	}
	s.lock.Unlock()
}

// ensures Tick() calles compile() and then resets the counter to step 1
func TestCompilationTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	// create state and give it a heartbeat to prevent it from pruning itself
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := s.newHeartbeat()
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash([]byte(hb.marshal()))
	s.heartbeats[s.participantIndex][heartbeatHash] = hb

	// remember entropy to verify that compile() gets called
	currentEntropy := s.currentEntropy

	// verify that tick is wrapping around properly
	s.currentStep = common.QuorumSize
	go s.tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)

	s.lock.Lock()
	if s.currentStep != 1 {
		t.Fatal("s.currentStep failed to roll over: ", s.currentStep)
	}

	// check if s.compile() got called
	if currentEntropy == s.currentEntropy {
		t.Fatal("Entropy did not change after tick wrapped around")
	}
	s.lock.Unlock()
}

// TestTickLock verifies that only one instance of Tick() can run at a time
func TestTickLock(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	// create state
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	// call tick twice
	go s.tick()
	go s.tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	// if two instances of Tick() are running, s.CurrentStep will update twice
	s.lock.Lock()
	if s.currentStep != 2 {
		t.Fatal("Double tick failed: ", s.currentStep)
	}
	s.lock.Unlock()
}
