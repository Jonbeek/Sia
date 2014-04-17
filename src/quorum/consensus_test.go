package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// test create heartbeat
func TestNewHeartbeat(t *testing.T) {
	s, err := CreateState(nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	hb, err := s.NewHeartbeat()
	if err != nil {
		t.Fatal(err)
	}

	// make sure that EntropyStage1 matches the hash of what gets stored
	storedEntropyHash, err := crypto.CalculateTruncatedHash(s.StoredEntropyStage2[:])
	if err != nil {
		t.Fatal(err)
	} else if hb.EntropyStage1 != storedEntropyHash {
		t.Fatal("NewHeartbeat() incorrectly producing EntropyStage1 from s.StoredEntropyStage2")
	}
}

// test heartbeat.marshal

// An incomplete set of tests: the more complete suite will attack the system
// as a whole.
func TestHandleSignedHeartbeat(t *testing.T) {
	// create a state and populate it with the signatories as participants
	s, err := CreateState(nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	// create some public keys
	pubKey1, secKey1, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	pubKey2, secKey2, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
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

func TestTossParticipant(t *testing.T) {
	// tossParticipant isn't yet implemented
}

func TestProcessHeartbeat(t *testing.T) {
	// create states and add them to each other
	s0, err := CreateState(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	s1, err := CreateState(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	s0.AddParticipant(s1.PublicKey, 1)
	s1.AddParticipant(s0.PublicKey, 0)

	// get the hash of the first heartbeat in s0
	var hash crypto.TruncatedHash
	for index := range s0.Heartbeats[0] {
		hash = index
	}

	// check that a valid heartbeat passes
	hb0 := s0.Heartbeats[0][hash]
	if err != nil {
		t.Fatal(err)
	}
	returnCode := s1.processHeartbeat(hb0, 0)
	if returnCode != 0 {
		t.Fatal("processHeartbeat threw out a valid heartbeat")
	}

	// check that invalid entropy fails
	// get the hash of the first hearbeat in s1
	for index := range s1.Heartbeats[1] {
		hash = index
	}
	hb1 := s1.Heartbeats[1][hash]

	// make heartbeat invalid (hb1.EntropyStage2 should be the 0 value)
	hb1.EntropyStage2[0] = 1
	returnCode = s0.processHeartbeat(hb1, 1)
	if returnCode != 1 {
		t.Fatal("processHeartbeat accepted an invalid heartbeat")
	}
}

func TestCompile(t *testing.T) {
	// Create states and add them to eachother as participants
	s0, err := CreateState(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	s1, err := CreateState(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := CreateState(nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	s3, err := CreateState(nil, 3)
	if err != nil {
		t.Fatal(err)
	}
	s0.AddParticipant(s1.PublicKey, 1)
	s0.AddParticipant(s2.PublicKey, 2)
	s0.AddParticipant(s3.PublicKey, 3)

	// fetch legal heartbeat for s1
	var hash crypto.TruncatedHash
	for index := range s1.Heartbeats[1] {
		hash = index
	}
	hb1 := s1.Heartbeats[1][hash]
	shb1, err := s1.SignHeartbeat(hb1)
	if err != nil {
		t.Fatal(err)
	}

	// fetch legal heartbeat for s3
	for index := range s3.Heartbeats[3] {
		hash = index
	}
	hb3a := s3.Heartbeats[3][hash]
	shb3a, err := s3.SignHeartbeat(hb3a)
	if err != nil {
		t.Fatal(err)
	}

	// create a second illegal heartbeat for s3
	var hb3b Heartbeat
	hb3b.EntropyStage1[0] = 0
	hb3b.EntropyStage2 = hb3a.EntropyStage2
	shb3b, err := s3.SignHeartbeat(&hb3b)
	if err != nil {
		t.Fatal(err)
	}

	// send the SignedHeartbeats to s0
	returnCode := s0.HandleSignedHeartbeat(shb1)
	if returnCode != 0 {
		t.Fatal("Expecting shb1 to be valid: ", returnCode)
	}
	returnCode = s0.HandleSignedHeartbeat(shb3a)
	if returnCode != 0 {
		t.Fatal("Expecting shb3a to be valid: ", returnCode)
	}
	returnCode = s0.HandleSignedHeartbeat(shb3b)
	if returnCode != 0 {
		t.Fatal("Expecting shb3b to be valid: ", returnCode)
	}

	s0.Compile()

	// check that hosts arrive at the same participantOrdering
	participantOrdering1 := s1.participantOrdering()
	participantOrdering2 := s2.participantOrdering()
	if participantOrdering1 != participantOrdering2 {
		t.Fatal("partcipantOrderings for s1 and s2 are not identical!")
	}

	// verify that upon processing, s1 is not thrown from s0, and is processed correctly
	if s0.Participants[1] == nil {
		t.Fatal("s1 thrown from s0 despite having a fair heartbeat")
	}

	// verify that upon processing, s2 is thrown from s0 (doesn't have heartbeat)
	if s0.Participants[2] != nil {
		t.Fatal("s2 not thrown from s0 despite having no heartbeats")
	}

	// verify that upon processing, s3 is thrown from s0 (too many heartbeats)
	if s0.Participants[3] != nil {
		t.Fatal("s3 not thrown from s0 despite having multiple heartbeats")
	}

	// verify that a new heartbeat was made, formatted into a SignedHeartbeat, and sent off
}

// Ensures that Tick() updates CurrentStep
func TestRegularTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	s, err := CreateState(nil, 0)
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

	s, err := CreateState(nil, 0)
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

func TestTickLock(t *testing.T) {
	// this is a long test
	if testing.Short() {
		t.Skip()
	}

	// create state
	s, err := CreateState(nil, 0)
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
