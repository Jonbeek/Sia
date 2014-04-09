package quorum

import (
	"common"
	"common/crypto"
	"common/log"
	"time"
)

// Part of the Byzantine Generals Problem
// SignedHeartbeat contains a heartbeat from a host
// which has been signed by the host, and then by
// each additional host that has seen it
type SignedHeartbeat struct {
	Heartbeat     *Heartbeat
	HeartbeatHash crypto.Hash
	Signatures    []crypto.Signature
	Signatories   []crypto.PublicKey
}

// Heartbeat contains all of the information that a host needs to
// participate in the quorum. This includes entropy proofs, file
// proofs, and transactions from hosts.
type Heartbeat struct {
	EntropyStage1 crypto.Hash
	EntropyStage2 common.Entropy
}

// Using the current State, NewHeartbeat creates a heartbeat that
// fulfills all of the requirements of the quorum.
//
// This function is incomplete
func (s *State) NewHeartbeat() (hb *Heartbeat) {
	return
}

// Checks that a heartbeat follows all rules, including
// proper stage 2 reveals.
func (hb *Heartbeat) IsValid() (rv bool) {
	if len(hb.EntropyStage2) != common.EntropyVolume {
		rv = false
		return
	}

	rv = true
	return
}

// HandleSignedHeartbeat takes a heartbeat that has been signed
// as a part of the concensus algorithm, and follows all the rules
// that are necessary to ensure that all honest hosts arrive at
// the same conclusions about the actions of their peers.
//
// See the paper 'The Byzantine Generals Problem' for more insight
// on the algorithms used here. Paper can be found in
// doc/The Byzantine Generals Problem
//
// Some of the logging in HandleSignedHeartbeat may be incomplete
//
// This function is called concurrently, mutexes will be needed when
// accessing or altering the State
//
// It is assumed that when this function is called, the Heartbeat in
// question will already be in memory, and was correctly signed by the
// first signatory, and that it matches its hash.
//
// Currently, host does not check if it's own signature is in the
// pile before adding its own signature again
//
// The return code is purely for the testing suite. The numbers are chosen
// arbitrarily
func (s *State) HandleSignedHeartbeat(sh *SignedHeartbeat) (returnCode int) {
	// Check that the slices of signatures and signatories are of the same length
	if len(sh.Signatures) != len(sh.Signatories) {
		log.Infoln("SignedHeartbeat has mismatched signatures")
		returnCode = 1
		return
	}

	// s.CurrentStep must be less than or equal to len(sh.Signatories), unless the
	// current step is common.QuorumSize and len(sh.Signatories) == 1
	if s.CurrentStep > len(sh.Signatories) {
		if s.CurrentStep == common.QuorumSize && len(sh.Signatories) == 1 {
			// sleep long enough to pass the first requirement
			time.Sleep(common.StepDuration)
			// now continue to rest of function
		} else {
			log.Infoln("Received an out-of-sync SignedHeartbeat")
			returnCode = 2
			return
		}
	}

	// Check that heartbeat is from a participant
	// All participants have a map in the heartbeat map
	_, exists := s.Heartbeats[sh.Signatories[0]]
	if !exists {
		log.Infoln("Received a heartbeat from a non-participant")
		returnCode = 3
		return
	}

	_, exists = s.Heartbeats[sh.Signatories[0]][sh.HeartbeatHash]
	if exists {
		// We already have this heartbeat, no action needed
		// this will happen frequently, no logging needed either
		returnCode = 8
		return
	}

	// while processing signatures, signedMessage will keep growing
	var signedMessage crypto.SignedMessage
	signedMessage.Message = string(sh.HeartbeatHash[:])
	// keep a map of which signatories have already been confirmed
	previousSignatories := make(map[crypto.PublicKey]bool)

	for i, signatory := range sh.Signatories {
		// Verify that the signatory is a participant in the quorum
		_, exists := s.Participants[signatory]
		if !exists {
			log.Infoln("Received a heartbeat signed by an invalid signatory")
			returnCode = 4
			return
		}

		// Verify that the signatory has only been seen once in the current SignedHeartbeat
		if previousSignatories[signatory] {
			log.Infoln("Received a double-signed heartbeat")
			returnCode = 5
			return
		}

		// record that we've seen this signatory in the current SignedHeartbeat
		previousSignatories[signatory] = true

		// verify the signature
		signedMessage.Signature = sh.Signatures[i]
		verification, err := crypto.Verify(signatory, signedMessage)
		if err != nil {
			log.Errorln(err)
			return
		}

		// throwing the signature into the message here makes code cleaner in the loop
		// and after we sign it to send it to everyone else
		signedMessage.Message = signedMessage.CombinedMessage()

		// check status of verification
		if !verification {
			log.Infoln("Received invalid signature in SignedHeartbeat")
			returnCode = 6
			return
		}
	}

	// Add heartbeat to list of seen heartbeats
	// Will add a signed heartbeat even if invalid
	s.Heartbeats[sh.Signatories[0]][sh.HeartbeatHash] = sh.Heartbeat

	// See that heartbeat is valid (correct parent, etc.)
	if !sh.Heartbeat.IsValid() {
		log.Infoln("Received an invalid heartbeat")
		returnCode = 7
		return
	}

	// Sign the stack of signatures and send it to all hosts
	_, err := crypto.Sign(s.SecretKey, signedMessage.Message)
	if err != nil {
		log.Errorln(err)
	}

	// Send the new message to everybody

	returnCode = 0
	return
}

// Tick() should only be called once, and should run in its own go thread
// Every common.SETPLENGTH, it updates the currentStep value.
// When the value flips from common.QuorumSize to 1, Tick() calls
// 	integrateHeartbeats()
func (s *State) Tick() {
	// Every common.StepDuration, advance the state stage
	ticker := time.Tick(common.StepDuration)
	for _ = range ticker {
		if s.CurrentStep == common.QuorumSize {
			// call logic to compile block
			s.CurrentStep = 1
		} else {
			s.CurrentStep += 1
		}
	}
}
