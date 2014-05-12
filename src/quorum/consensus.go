package quorum

import (
	"bytes"
	"common"
	"common/crypto"
	"common/log"
	"encoding/gob"
	"errors"
	"fmt"
	"time"
)

// All information that needs to be passed between participants each block
type heartbeat struct {
	entropyStage1 crypto.TruncatedHash
	entropyStage2 common.Entropy
}

// Contains a heartbeat that has been signed iteratively, is a key part of the
// signed solution to the Byzantine Generals Problem
type SignedHeartbeat struct {
	heartbeat     *heartbeat
	heartbeatHash crypto.TruncatedHash
	signatories   []byte             // a list of everyone who's seen the heartbeat
	signatures    []crypto.Signature // their corresponding signatures
}

// Using the current State, newHeartbeat() creates a heartbeat that fulfills all
// of the requirements of the quorum.
func (s *State) newHeartbeat() (hb *heartbeat, err error) {
	hb = new(heartbeat)

	// Fetch value used to produce EntropyStage1 in prev. heartbeat
	hb.entropyStage2 = s.storedEntropyStage2

	// Generate EntropyStage2 for next heartbeat
	entropy, err := crypto.RandomByteSlice(common.EntropyVolume)
	if err != nil {
		return
	}
	copy(s.storedEntropyStage2[:], entropy) // convert entropy from slice to byte array

	// Use EntropyStage2 to generate EntropyStage1 for this heartbeat
	hb.entropyStage1, err = crypto.CalculateTruncatedHash(s.storedEntropyStage2[:])
	if err != nil {
		return
	}

	// more code will be added here

	return
}

// Convert heartbeat to []byte
func (hb *heartbeat) GobEncode() (gobHeartbeat []byte, err error) {
	// test for bad input
	if hb == nil {
		err = fmt.Errorf("Cannot Encode a nil object")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(hb.entropyStage1)
	if err != nil {
		return
	}
	err = encoder.Encode(hb.entropyStage2)
	if err != nil {
		return
	}

	gobHeartbeat = w.Bytes()
	return
}

// Convert []byte to heartbeat
func (hb *heartbeat) GobDecode(gobHeartbeat []byte) (err error) {
	r := bytes.NewBuffer(gobHeartbeat)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&hb.entropyStage1)
	if err != nil {
		return
	}
	err = decoder.Decode(&hb.entropyStage2)
	if err != nil {
		return
	}
	return
}

// take new heartbeat (our own), sign it, and package it into a signedHearteat
// I'm pretty sure this only follows a newHeartbeat() call; they can be merged
func (s *State) signHeartbeat(hb *heartbeat) (sh *SignedHeartbeat, err error) {
	sh = new(SignedHeartbeat)

	// confirm heartbeat and hash
	sh.heartbeat = hb
	gobHb, err := hb.GobEncode()
	if err != nil {
		return
	}
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash(gobHb)
	if err != nil {
		return
	}

	// fill out signatures
	sh.signatures = make([]crypto.Signature, 1)
	signedHb, err := s.secretKey.Sign(sh.heartbeatHash[:])
	if err != nil {
		return
	}
	sh.signatures[0] = signedHb.Signature
	sh.signatories = make([]byte, 1)
	sh.signatories[0] = s.self.index
	return
}

// Takes a signed heartbeat and broadcasts it to the quorum
func (s *State) announceSignedHeartbeat(sh *SignedHeartbeat) (err error) {
	s.broadcast(&common.Message{
		Proc: "State.HandleSignedHeartbeat",
		Args: *sh,
		Resp: nil,
	})
	return
}

var hsherrMismatchedSignatures = errors.New("SignedHeartbeat has mismatches signatures to signatories")
var hsherrOversigned = errors.New("Received an over-signed signedHeartbeat")
var hsherrNoSync = errors.New("Received an out-of-sync SignedHeartbeat")
var hsherrBounds = errors.New("Received an out of bounds index for signatory")
var hsherrNonParticipant = errors.New("Received heartbeat from non-participant")
var hsherrHaveHeartbeat = errors.New("Already have this heartbeat")
// incomplete, will finish with remainder of testing

// HandleSignedHeartbeat takes the payload of an incoming message of type
// 'incomingSignedHeartbeat' and verifies it according to the specification
//
// What sort of input error checking is needed for this function?
func (s *State) HandleSignedHeartbeat(sh SignedHeartbeat, arb *struct{}) error {
	// Check that the slices of signatures and signatories are of the same length
	if len(sh.signatures) != len(sh.signatories) {
		return hsherrMismatchedSignatures
	}

	// check that there are not too many signatures and signatories
	if len(sh.signatories) > common.QuorumSize {
		return hsherrOversigned
	}

	s.stepLock.Lock() // prevents a benign race condition; is here to follow best practices
	currentStep := s.currentStep
	s.stepLock.Unlock()
	// s.CurrentStep must be less than or equal to len(sh.Signatories), unless
	// there is a new block and s.CurrentStep is common.QuorumSize
	if currentStep > len(sh.signatories) {
		if currentStep == common.QuorumSize && len(sh.signatories) == 1 {
			// by waiting common.StepDuration, the new block will be compiled
			time.Sleep(common.StepDuration)
			// now continue to rest of function
		} else {
			return hsherrNoSync
		}
	}

	// Check bounds on first signatory
	if int(sh.signatories[0]) >= common.QuorumSize {
		return hsherrBounds
	}

	// we are starting to read from memory, initiate locks
	s.participantsLock.RLock()
	s.heartbeatsLock.Lock()
	defer s.participantsLock.RUnlock()
	defer s.heartbeatsLock.Unlock()

	// check that first signatory is a participant
	if s.participants[sh.signatories[0]] == nil {
		return hsherrNonParticipant
	}

	// Check if we have already received this heartbeat
	_, exists := s.heartbeats[sh.signatories[0]][sh.heartbeatHash]
	if exists {
		return hsherrHaveHeartbeat
	}

	// Check if we already have two heartbeats from this host
	if len(s.heartbeats[sh.signatories[0]]) >= 2 {
		log.Infoln("Received many invalid heartbeats from one host")
		return fmt.Errorf("Received many invalid heartbeats from one host")
	}

	// iterate through the signatures and make sure each is legal
	var signedMessage crypto.SignedMessage // grows each iteration
	signedMessage.Message = sh.heartbeatHash[:]
	previousSignatories := make(map[byte]bool) // which signatories have already signed
	for i, signatory := range sh.signatories {
		// Check bounds on the signatory
		if int(signatory) >= common.QuorumSize {
			log.Infoln("Received an out of bounds index")
			return fmt.Errorf("Received an out of bounds index")
		}

		// Verify that the signatory is a participant in the quorum
		if s.participants[signatory] == nil {
			log.Infoln("Received a heartbeat signed by an invalid signatory")
			return fmt.Errorf("Received a heartbeat signed by an invalid signatory")
		}

		// Verify that the signatory has only been seen once in the current SignedHeartbeat
		if previousSignatories[signatory] {
			log.Infoln("Received a double-signed heartbeat")
			return fmt.Errorf("Received a double-signed heartbeat")
		}

		// record that we've seen this signatory in the current SignedHeartbeat
		previousSignatories[signatory] = true

		// verify the signature
		signedMessage.Signature = sh.signatures[i]
		verification := s.participants[signatory].publicKey.Verify(&signedMessage)

		// check status of verification
		if !verification {
			log.Infoln("Received invalid signature in SignedHeartbeat")
			return fmt.Errorf("Received invalid signature in SignedHeartbeat")
		}

		// throwing the signature into the message here makes code cleaner in the loop
		// and after we sign it to send it to everyone else
		newMessage, err := signedMessage.CombinedMessage()
		signedMessage.Message = newMessage
		if err != nil {
			log.Infoln("Error while combining a signed message")
			return err
		}
	}

	// Add heartbeat to list of seen heartbeats
	s.heartbeats[sh.signatories[0]][sh.heartbeatHash] = sh.heartbeat

	// Sign the stack of signatures and send it to all hosts
	signedMessage, err := s.secretKey.Sign(signedMessage.Message)
	if err != nil {
		log.Fatalln(err)
	}

	// add our signature to the signedHeartbeat
	sh.signatures = append(sh.signatures, signedMessage.Signature)
	sh.signatories = append(sh.signatories, s.self.index)

	// broadcast the message to the quorum
	err = s.announceSignedHeartbeat(&sh)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	return nil
}

func (sh *SignedHeartbeat) GobEncode() (gobSignedHeartbeat []byte, err error) {
	// error check the input
	if sh == nil {
		err = fmt.Errorf("Cannot encode a nil object")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(sh.heartbeat)
	if err != nil {
		return
	}
	err = encoder.Encode(sh.heartbeatHash)
	if err != nil {
		return
	}
	err = encoder.Encode(sh.signatories)
	if err != nil {
		return
	}
	err = encoder.Encode(sh.signatures)
	if err != nil {
		return
	}

	gobSignedHeartbeat = w.Bytes()
	return
}

func (shb *SignedHeartbeat) GobDecode(gobSignedHeartbeat []byte) (err error) {
	if gobSignedHeartbeat == nil {
		err = fmt.Errorf("cannot decode a nil byte slice")
		return
	}

	r := bytes.NewBuffer(gobSignedHeartbeat)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&shb.heartbeat)
	if err != nil {
		return
	}
	err = decoder.Decode(&shb.heartbeatHash)
	if err != nil {
		return
	}
	err = decoder.Decode(&shb.signatories)
	if err != nil {
		return
	}
	err = decoder.Decode(&shb.signatures)
	if err != nil {
		return
	}

	return
}

// participants are processed in a random order each block, determined by the
// entropy for the block. participantOrdering() deterministically picks that
// order, using entropy from the state.
func (s *State) participantOrdering() (participantOrdering [common.QuorumSize]byte) {
	// create an in-order list of participants
	for i := range participantOrdering {
		participantOrdering[i] = byte(i)
	}

	// shuffle the list of participants
	for i := range participantOrdering {
		newIndex, err := s.randInt(i, common.QuorumSize)
		if err != nil {
			log.Fatalln(err)
		}
		tmp := participantOrdering[newIndex]
		participantOrdering[newIndex] = participantOrdering[i]
		participantOrdering[i] = tmp
	}

	return
}

// Removes all traces of a participant from the State
func (s *State) tossParticipant(pi byte) {
	// remove from s.Participants
	s.participants[pi] = nil

	// remove from s.PreviousEntropyStage1
	var emptyEntropy common.Entropy
	zeroHash, err := crypto.CalculateTruncatedHash(emptyEntropy[:])
	if err != nil {
		log.Fatal(err)
	}
	s.previousEntropyStage1[pi] = zeroHash

	// nil map in s.Heartbeats
	s.heartbeats[pi] = nil
}

// Update the state according to the information presented in the heartbeat
// processHeartbeat uses return codes for testing purposes
func (s *State) processHeartbeat(hb *heartbeat, i byte) int {
	// compare EntropyStage2 to the hash from the previous heartbeat
	expectedHash, err := crypto.CalculateTruncatedHash(hb.entropyStage2[:])
	if err != nil {
		log.Fatalln(err)
	}
	if expectedHash != s.previousEntropyStage1[i] {
		s.tossParticipant(i)
		return 1
	}

	print("Confirming Participant ")
	println(i)

	// Add the EntropyStage2 to UpcomingEntropy
	th, err := crypto.CalculateTruncatedHash(append(s.upcomingEntropy[:], hb.entropyStage2[:]...))
	s.upcomingEntropy = common.Entropy(th)

	// store entropyStage1 to compare with next heartbeat from this participant
	s.previousEntropyStage1[i] = hb.entropyStage1

	return 0
}

// compile() takes the list of heartbeats and uses them to advance the state.
func (s *State) compile() {
	// fetch a participant ordering
	participantOrdering := s.participantOrdering()

	// Lock down s.participants and s.heartbeats for editing
	s.participantsLock.Lock()
	s.heartbeatsLock.Lock()

	// Read heartbeats, process them, then archive them.
	for _, participant := range participantOrdering {
		if s.participants[participant] == nil {
			continue
		}

		// each participant must submit exactly 1 heartbeat
		if len(s.heartbeats[participant]) != 1 {
			s.tossParticipant(participant)
			continue
		}

		// this is the only way I know to access the only element of a map;
		// the key is unknown
		for _, hb := range s.heartbeats[participant] {
			s.processHeartbeat(hb, participant)
		}

		// archive heartbeats (unimplemented)

		// clear heartbeat list for next block
		s.heartbeats[participant] = make(map[crypto.TruncatedHash]*heartbeat)
	}

	s.participantsLock.Unlock()
	s.heartbeatsLock.Unlock()

	// move UpcomingEntropy to CurrentEntropy
	s.currentEntropy = s.upcomingEntropy

	// generate, sign, and announce new heartbeat
	hb, err := s.newHeartbeat()
	if err != nil {
		log.Fatalln(err)
	}
	shb, err := s.signHeartbeat(hb)
	if err != nil {
		log.Fatalln(err)
	}
	err = s.announceSignedHeartbeat(shb)
	if err != nil {
		log.Fatalln(err)
	}
}

// Tick() updates s.CurrentStep, and calls compile() when all steps are complete
func (s *State) tick() {
	// Every common.StepDuration, advance the state stage
	ticker := time.Tick(common.StepDuration)
	for _ = range ticker {
		s.stepLock.Lock()
		if s.currentStep == common.QuorumSize {
			println("compiling")
			s.compile()
			s.currentStep = 1
		} else {
			println("stepping")
			s.currentStep += 1
		}
		s.stepLock.Unlock()
	}
}
