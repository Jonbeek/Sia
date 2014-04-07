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
	Heartbeat   Heartbeat
	Signatures  []string
	Signatories []crypto.PublicKey
}

// Heartbeat contains all of the information that a host needs to
// participate in the quorum. This includes entropy proofs, file
// proofs, and transactions from hosts.
type Heartbeat struct {
	EntropyStage1 string
	EntropyStage2 string
}

// Using the current State, NewHeartbeat creates a heartbeat that
// fulfills all of the requirements of the quorum.
func (s *State) NewHeartbeat(h Heartbeat) {
	h.EntropyStage1 = "tbi"
	h.EntropyStage2 = "toBeImplemented"
}

// Checks if this heartbeat is the zero value
func (hb *Heartbeat) IsEmpty() (rv bool) {
	if hb.EntropyStage1 != "" {
		rv = false
		return
	}

	if hb.EntropyStage2 != "" {
		rv = false
		return
	}

	rv = true
	return
}

// Check if the two heartbeats are identical
func (first *Heartbeat) IsEqual(second *Heartbeat) (rv bool) {
	if first.EntropyStage1 != second.EntropyStage1 {
		rv = false
		return
	}

	if first.EntropyStage2 != second.EntropyStage2 {
		rv = false
		return
	}

	rv = true
	return
}

func (hb *Heartbeat) IsValid() (rv bool) {
	if len(hb.EntropyStage2) != common.ENTROPYVOLUME {
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
// Some of the logging in this function may be incomplete
//
// I am uncertain if this function will be called concurrently
func (s *State) HandleSignedHeartbeat(sh *SignedHeartbeat) {
	// Check that the slices of signatures and signatories are of the same length
	if len(sh.Signatures) != len(sh.Signatories) {
		log.Infoln("SignedHeartbeat has mismatched signatures")
		return
	}

	// Check that there is at least 1 signatory
	if len(sh.Signatories) == 0 {
		log.Infoln("Reveiced an unsigned SignedHeartbeat")
		return
	}

	// s.CurrentStep must be less than or equal to len(sh.Signatories), unless the
	// current step is 180 and len(sh.Signatories) == 1
	if s.CurrentStep > len(sh.Signatories) && !(s.CurrentStep == 180 && len(sh.Signatories) == 1) {
		log.Infoln("Received an invalid SignedHeartbeat")
		return
	}

	// See if the heartbeat we got is the zero value
	if sh.Heartbeat.IsEmpty() {
		log.Infoln("Received an empty SignedHeartbeat")
		return
	}

	// Check if we have this heartbeat
	if sh.Heartbeat.IsEqual(s.Heartbeats[sh.Signatories[0]]) {
		// no logging needed, this will happen frequently
		return
	}

	var previousSignatories map[crypto.PublicKey]bool

	for _, signatory := range sh.Signatories {
		// if s.CurrentStep == 180 && len(sh.Signatories == 1, then we need to store the new
		// heartbeat as a part of next round, not sure the best way to do that

		// Verify that the signatory is a participant in the quorum
		if s.Participants[signatory].IsEmpty() {
			log.Infoln("Received a heartbeat from an invalid signatory")
			return
		}

		// Verify that the signatory has only been seen once in the current SignedHeartbeat
		if previousSignatories[signatory] {
			log.Infoln("Received a double-signed heartbeat")
			return
		}

		// signal that we've seen this signatory in the current SignedHeartbeat
		previousSignatories[signatory] = true

		// See if the current signature is valid
	}

	// See that heartbeat is valid (correct parent, etc.)
	if !sh.Heartbeat.IsValid() {
		log.Infoln("Received an invalid heartbeat")
		return
	}

	// Add to list of Heartbeats
	s.Heartbeats[sh.Signatories[0]] = &sh.Heartbeat

	// After adding/getting a new heartbeat, you need to add your signature
	// and send it to everybody else

	// If at any point a host is outed as having been dishonest,
	// leave some sort of warning value in the map
}

// Tick() should only be called once, and should run in its own go thread
// Every common.SETPLENGTH, it updates the currentStep value.
// When the value flips from common.QUORUMSIZE to 1, Tick() calls
// 	integrateHeartbeats()
func (s *State) Tick() {
	// Every common.STEPLENGTH, advance the state stage
	ticker := time.Tick(common.STEPLENGTH)
	for _ = range ticker {
		if s.CurrentStep == common.QUORUMSIZE {
			// call logic to compile block
			s.CurrentStep = 1
		} else {
			s.CurrentStep += 1
		}
	}
}
