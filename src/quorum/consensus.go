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
	HeartbeatHash string
	Signatures    []string
	Signatories   []crypto.PublicKey
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

// Checks that a heartbeat follows all rules, including
// proper stage 2 reveals.
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
// Some of the logging in HandleSignedHeartbeat may be incomplete
//
// This function is called concurrently, mutexes will be needed when
// accessing or altering the State
//
// It is assumed that when this function is called, the Heartbeat in
// question will already be in memory, and was correctly signed by the
// first signatory
func (s *State) HandleSignedHeartbeat(sh *SignedHeartbeat) {
	// Check that the slices of signatures and signatories are of the same length
	if len(sh.Signatures) != len(sh.Signatories) {
		log.Infoln("SignedHeartbeat has mismatched signatures")
		return
	}

	// s.CurrentStep must be less than or equal to len(sh.Signatories), unless the
	// current step is common.QUORUMSIZE and len(sh.Signatories) == 1
	if s.CurrentStep > len(sh.Signatories) && !(s.CurrentStep == common.QUORUMSIZE && len(sh.Signatories) == 1) {
		log.Infoln("Received an invalid SignedHeartbeat")
		return
	}

	if s.CurrentStep > len(sh.Signatories) {
		if s.CurrentStep == common.QUORUMSIZE && len(sh.Signatories) == 1 {
			// sleep long enough to pass the first requirement
			time.Sleep(common.STEPLENGTH)

			// verify that the first requirement still holds
			// this is really just a debugging statement
			if s.CurrentStep > len(sh.Signatories) {
				log.Errorln("Incorrect HandleSignedHeartbeat logic!")
				// maybe also log more information
				return
			}

			// now continue to rest of function
		} else {
			log.Infoln("Received an out-of-sync SignedHeartbeat")
			return
		}
	}

	// See if the heartbeat we got is the zero value
	// A zero valued heartbeat is equivalent to no heartbeat
	if sh.Heartbeat.IsEmpty() {
		log.Infoln("Received an empty SignedHeartbeat")
		return
	}

	// Check if we have this heartbeat
	if sh.Heartbeat.IsEqual(s.Heartbeats[sh.Signatories[0]]) {
		// no logging needed, this will happen frequently
		return
	}

	// while processing signatures, signedMessage will keep growing
	// additionally, we will keep a map of which signatures we have
	// already seen in this SignedHeartbeat
	signedMessage := sh.HeartbeatHash
	var previousSignatories map[crypto.PublicKey]bool

	for i, signatory := range sh.Signatories {
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

		// record that we've seen this signatory in the current SignedHeartbeat
		previousSignatories[signatory] = true

		// verify the signature
		// this append inefficiency is enough to make me consider avoiding strings
		signedMessage = string(append([]byte(signedMessage), sh.Signatures[i]...))
		verification, err := crypto.Verify(string(signatory), signedMessage)

		// check error message
		if err != nil {
			log.Infoln(err)
			return
		}

		// check status of verification
		if !verification {
			log.Infoln("Received invalid signature in SignedHeartbeat")
			return
		}
	}

	// See that heartbeat is valid (correct parent, etc.)
	if !sh.Heartbeat.IsValid() {
		log.Infoln("Received an invalid heartbeat")
		return
	}

	// check if we have a different heartbeat for this host
	// we already know that we don't have the same heartbeat
	if !s.Heartbeats[sh.Signatories[0]].IsEmpty() {
		// remove host from Participants, charge him fines, etc.
		// this process will probably be handled by 'indictments.go'
		// for the time being, that file will not exist
	} else {
		// Add to list of Heartbeats
		// A conflict heartbeat is not added, which could cause
		// DDOS related problems
		//
		// this map solution needs greater consideration
		s.Heartbeats[sh.Signatories[0]] = sh.Heartbeat
	}

	// Sign the stack of signatures and send it to all hosts
	_, err := crypto.Sign(string(s.SecretKey), signedMessage)
	if err != nil {
		log.Errorln(err)
	}

	// Send the new message to everybody
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
