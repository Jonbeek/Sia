package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// Verify that NewHeartbeat() produces valid heartbeats
func TestNewHeartbeat(t *testing.T) {
	// create a state, and then a heartbeat
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := s.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}

	// verify that entropy is being properly generated when making the heartbeat
	storedEntropyHash, err := crypto.CalculateTruncatedHash(s.StoredEntropyStage2[:])
	if err != nil {
		t.Fatal(err)
	} else if hb.EntropyStage1 != storedEntropyHash {
		t.Fatal("NewHeartbeat() incorrectly producing EntropyStage1 from s.StoredEntropyStage2")
	}
}

// Marshalling and Unmarshalling should result in equivalent Heartbeats
func TestHeartbeatMarshalling(t *testing.T) {
	s, err := CreateState(common.NewZeroNetwork(), 0)
	if err != nil {
		t.Fatal(err)
	}

	// verify that a heartbeat, once unmarshalled, is identical to the original
	hbOriginal, err := s.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	hbMarshalled := hbOriginal.Marshal()
	hbUnmarshalled, err := UnmarshalHeartbeat(hbMarshalled)
	if err != nil {
		t.Fatal(err)
	}
	if hbOriginal.EntropyStage1 != hbUnmarshalled.EntropyStage1 {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}
	if hbOriginal.EntropyStage2 != hbUnmarshalled.EntropyStage2 {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}

	// verify that input is being checked for UnmarshalHeartbeat
	_, err = UnmarshalHeartbeat(hbMarshalled[1:])
	if err == nil {
		t.Fatal("Heartbeat unmarshalling succeded with a short input")
	}
	_, err = UnmarshalHeartbeat(append(hbMarshalled, hbMarshalled...))
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
	hb, err := s.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	originalSignedHeartbeat, err := s.SignHeartbeat(hb)
	if err != nil {
		t.Fatal(err)
	}
	marshalledSH, err := originalSignedHeartbeat.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	unmarshalledSH, err := UnmarshalSignedHeartbeat(marshalledSH)
	if string(unmarshalledSH.Heartbeat.Marshal()) != string(originalSignedHeartbeat.Heartbeat.Marshal()) {
		t.Fatal("Heartbeat changed after being marshalled and unmarshalled")
	}
	if unmarshalledSH.HeartbeatHash != originalSignedHeartbeat.HeartbeatHash {
		t.Fatal("HeartbeatHash changed after being marshalled and unmarshalled")
	}
	if len(unmarshalledSH.Signatures) != len(originalSignedHeartbeat.Signatures) {
		t.Fatal("Length of SignedHeartbeat.Signatures changed after being marshalled and unmarshalled")
	}
	if len(unmarshalledSH.Signatories) != len(originalSignedHeartbeat.Signatories) {
		t.Fatal("Length of SignedHeartbeat.Signatories changed after being marshalled and unmarshalled")
	}
	for i := 0; i < len(unmarshalledSH.Signatures); i++ {
		if unmarshalledSH.Signatures[i] != originalSignedHeartbeat.Signatures[i] {
			t.Fatal("For i=", i, ", unmarshalledSH.Signatures[i] did not equal originalSignedHeartbeat.Signatures[i]")
		}
		if unmarshalledSH.Signatories[i] != originalSignedHeartbeat.Signatories[i] {
			t.Fatal("For i=", i, ", unmarshalledSH.Signatories[i] did not equal originalSignedHeartbeat.Signatories[i]")
		}
	}

	// verify that input is being checked for SignedHeartbeat.Marshal()
	originalSignedHeartbeat.Signatures = make([]crypto.Signature, 2*common.QuorumSize)
	_, err = originalSignedHeartbeat.Marshal()
	if err == nil {
		t.Fatal("SignedHeartbeat.Marshal() needs to check the length of msh.Signatures")
	}
	originalSignedHeartbeat.Signatures = make([]crypto.Signature, 2)
	if err == nil {
		t.Fatal("SignedHeartbeat.Marshal() needs to check that msh.Signatures and msh.Signatories are equal in length")
	}

	// verify that input is being checked for UnmarshalSignedHeartbeat()
	_, err = UnmarshalSignedHeartbeat(marshalledSH[:MarshalledHeartbeatLen()])
	if err == nil {
		t.Fatal("UnmarshalSignedHeartbeat not checking input length")
	}
	_, err = UnmarshalSignedHeartbeat(marshalledSH[1:])
	if err == nil {
		t.Fatal("UnmarshalSignedHeartbeat succeeded when input was too short")
	}
	_, err = UnmarshalSignedHeartbeat(append(marshalledSH, marshalledSH...))
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
	msh, err := sh.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode := s.HandleSignedHeartbeat(msh)
	if returnCode != 0 {
		t.Fatal("expected heartbeat to succeed:", returnCode)
	}

	// verify that a repeat heartbeat gets ignored
	msh, err = sh.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.HandleSignedHeartbeat(msh)
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
	msh, err = sh.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.HandleSignedHeartbeat(msh)
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
	msh, err = sh.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.HandleSignedHeartbeat(msh)
	if returnCode != 5 {
		t.Fatal("expected heartbeat to be rejected for duplicate signatures: ", returnCode)
	}

	// remove second signature
	sh.Signatures = sh.Signatures[:1]
	sh.Signatories = sh.Signatories[:1]

	// handle heartbeat when tick is larger than num signatures
	s.CurrentStep = 2
	msh, err = sh.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s.HandleSignedHeartbeat(msh)
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
		msh, err = sh.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		returnCode = s.HandleSignedHeartbeat(msh)
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
	hb0, err := s0.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	returnCode := s1.processHeartbeat(hb0, 0)
	if returnCode != 0 {
		t.Fatal("processHeartbeat threw out a valid heartbeat")
	}

	// check that invalid entropy fails
	hb1, err := s1.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	hb1.EntropyStage2[0] = 1
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
	hb0, err := s0.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	shb0, err := s0.SignHeartbeat(hb0)
	if err != nil {
		t.Fatal(err)
	}

	// fetch legal heartbeat for s2
	hb2a, err := s2.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	shb2a, err := s2.SignHeartbeat(hb2a)
	if err != nil {
		t.Fatal(err)
	}

	// create a second illegal heartbeat for s2
	var hb2b Heartbeat
	hb2b.EntropyStage2 = hb2a.EntropyStage2
	shb2b, err := s2.SignHeartbeat(&hb2b)
	if err != nil {
		t.Fatal(err)
	}

	// send the SignedHeartbeats to s0
	mshb0, err := shb0.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode := s0.HandleSignedHeartbeat(mshb0)
	if returnCode != 0 {
		t.Fatal("Expecting shb0 to be valid: ", returnCode)
	}
	mshb2a, err := shb2a.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s0.HandleSignedHeartbeat(mshb2a)
	if returnCode != 0 {
		t.Fatal("Expecting shb2a to be valid: ", returnCode)
	}
	mshb2b, err := shb2b.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	returnCode = s0.HandleSignedHeartbeat(mshb2b)
	if returnCode != 0 {
		t.Fatal("Expecting shb2b to be valid: ", returnCode)
	}

	s0.Compile()

	// check that hosts arrive at the same participantOrdering
	participantOrdering1 := s1.participantOrdering()
	participantOrdering2 := s2.participantOrdering()
	if participantOrdering1 != participantOrdering2 {
		t.Fatal("partcipantOrderings for s1 and s2 are not identical!")
	}

	// verify that upon processing, s0 is not thrown from s0, and is processed correctly
	if s0.Participants[0] == nil {
		t.Fatal("s0 thrown from s0 despite having a fair heartbeat")
	}

	// verify that upon processing, s1 is thrown from s0 (doesn't have heartbeat)
	if s0.Participants[1] != nil {
		t.Fatal("s1 not thrown from s0 despite having no heartbeats")
	}

	// verify that upon processing, s3 is thrown from s0 (too many heartbeats)
	if s0.Participants[2] != nil {
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
	s.CurrentStep = 1
	go s.Tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	if s.CurrentStep != 2 {
		t.Fatal("s.CurrentStep failed to update correctly: ", s.CurrentStep)
	}
}

// ensures Tick() calles Compile() and then resets the counter to step 1
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
	hb, err := s.NewHeartbeat()
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash([]byte(hb.Marshal()))
	s.Heartbeats[s.ParticipantIndex][heartbeatHash] = hb

	// remember entropy to verify that compile() gets called
	currentEntropy := s.CurrentEntropy

	// verify that tick is wrapping around properly
	s.CurrentStep = common.QuorumSize
	go s.Tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	if s.CurrentStep != 1 {
		t.Fatal("s.CurrentStep failed to roll over: ", s.CurrentStep)
	}

	// check if s.Compile() got called
	if currentEntropy == s.CurrentEntropy {
		t.Fatal("Entropy did not change after tick wrapped around")
	}
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
	go s.Tick()
	go s.Tick()
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	// if two instances of Tick() are running, s.CurrentStep will update twice
	if s.CurrentStep != 2 {
		t.Fatal("Double tick failed: ", s.CurrentStep)
	}
}
