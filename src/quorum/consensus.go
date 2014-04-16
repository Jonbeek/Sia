package quorum

import (
	"common"
	"common/crypto"
	"common/log"
	"time"
)

// Heartbeat contains all of the information that a host needs to
// participate in the quorum. This includes entropy proofs, file
// proofs, and transactions from hosts.
type Heartbeat struct {
	EntropyStage1 crypto.TruncatedHash
	EntropyStage2 common.Entropy
}

// Part of the Byzantine Generals Problem
// SignedHeartbeat contains a heartbeat from a host
// which has been signed by the host, and then by
// each additional host that has seen it
type SignedHeartbeat struct {
	Heartbeat     *Heartbeat
	HeartbeatHash crypto.TruncatedHash
	Signatures    []crypto.Signature
	Signatories   []ParticipantIndex
}

// Using the current State, NewHeartbeat creates a heartbeat that
// fulfills all of the requirements of the quorum.
func (s *State) NewHeartbeat() (hb *Heartbeat, err error) {
	var heartbeat Heartbeat
	hb = &heartbeat
	// Fetch value used to produce EntropyStage1 in prev. heartbeat
	hb.EntropyStage2 = s.StoredEntropyStage2

	// Generate EntropyStage2 for next heartbeat
	rawEntropy, err := crypto.RandomByteSlice(common.EntropyVolume)
	if err != nil {
		return
	}
	copy(s.StoredEntropyStage2[:], rawEntropy)

	// Use EntropyStage2 to generate EntropyStage1 for this heartbeat
	hb.EntropyStage1, err = crypto.CalculateTruncatedHash(s.StoredEntropyStage2[:])
	if err != nil {
		return
	}

	// more code will be added here

	return
}

// Convert Heartbeat to string
func (hb *Heartbeat) Marshal() (marshalledHeartbeat string) {
	marshalledHeartbeat = string(append(hb.EntropyStage1[:], hb.EntropyStage2[:]...))
	return
}

// func UnmarshalHeartbeat(marshalledHeartbeat string) {
// }

func (s *State) SignHeartbeat(hb *Heartbeat) (sh *SignedHeartbeat, err error) {
	var signedHeartbeat SignedHeartbeat
	sh = &signedHeartbeat

	// confirm heartbeat and hash
	sh.Heartbeat = hb
	marshalledHb := hb.Marshal()
	sh.HeartbeatHash, err = crypto.CalculateTruncatedHash([]byte(marshalledHb))
	if err != nil {
		return
	}

	// fill out sigantures
	sh.Signatures = make([]crypto.Signature, 1)
	signedHb, err := crypto.Sign(s.SecretKey, string(sh.HeartbeatHash[:]))
	if err != nil {
		return
	}
	sh.Signatures[0] = signedHb.Signature
	sh.Signatories = make([]ParticipantIndex, 1)
	sh.Signatories[0] = s.ParticipantIndex
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
// This function is called concurrently, mutexes will be needed when
// accessing or altering the State
//
// It is assumed that when this function is called, the Heartbeat in
// question will already be in memory, and was correctly signed by the
// first signatory, the the first signatory is a participant, and that
// it matches its hash. And that the first signatory is used to store
// the heartbeat
//
// The return code is purely for the testing suite. The numbers are chosen
// arbitrarily
//
// Really, the HandleSignedHeartbeat should take a string, and open it up
// all the way from that point. Which means the test suite needs updating
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

	// Check if we have already received this heartbeat
	_, exists := s.Heartbeats[sh.Signatories[0]][sh.HeartbeatHash]
	if exists {
		returnCode = 8
		return
	}

	// while processing signatures, signedMessage will keep growing
	var signedMessage crypto.SignedMessage
	signedMessage.Message = string(sh.HeartbeatHash[:])
	// keep a map of which signatories have already been confirmed
	previousSignatories := make(map[ParticipantIndex]bool)

	for i, signatory := range sh.Signatories {
		// Verify that the signatory is a participant in the quorum
		if s.Participants[signatory] == nil {
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
		verification, err := crypto.Verify(s.Participants[signatory].PublicKey, signedMessage)
		if err != nil {
			log.Fatalln(err)
			return
		}

		// check status of verification
		if !verification {
			log.Infoln("Received invalid signature in SignedHeartbeat")
			returnCode = 6
			return
		}

		// throwing the signature into the message here makes code cleaner in the loop
		// and after we sign it to send it to everyone else
		signedMessage.Message = signedMessage.CombinedMessage()
	}

	// Add heartbeat to list of seen heartbeats
	// Don't check if heartbeat is valid, that's for Compile()
	s.Heartbeats[sh.Signatories[0]][sh.HeartbeatHash] = sh.Heartbeat

	// Sign the stack of signatures and send it to all hosts
	_, err := crypto.Sign(s.SecretKey, signedMessage.Message)
	if err != nil {
		log.Fatalln(err)
	}

	// Send the new message to everybody

	returnCode = 0
	return
}

func (s *State) tossParticipant(pi ParticipantIndex) {
	// remove from s.Participants
	s.Participants[pi] = nil

	// remove from s.PreviousEntropyStage1
	var emptyEntropy common.Entropy
	zeroHash, err := crypto.CalculateTruncatedHash(emptyEntropy[:])
	s.PreviousEntropyStage1[pi] = zeroHash
	if err != nil {
		log.Fatal(err)
	}

	// nil map in s.Heartbeats
	s.Heartbeats[pi] = nil
}

func (s *State) processHeartbeat(hb *Heartbeat, i ParticipantIndex) int {
	// compare EntropyStage2 to the hash from the previous heartbeat
	expectedHash, err := crypto.CalculateTruncatedHash(hb.EntropyStage2[:])
	if err != nil {
		log.Fatalln(err)
	}
	if expectedHash != s.PreviousEntropyStage1[i] {
		s.tossParticipant(i)
		return 1
	}

	// Add the EntropyStage2 to UpcomingEntropy
	th, err := crypto.CalculateTruncatedHash(append(s.UpcomingEntropy[:], hb.EntropyStage2[:]...))
	s.UpcomingEntropy = common.Entropy(th)
	// update PreviousEntropy, to compare this EntropyStage1 against the next
	// EntropyStage1
	s.PreviousEntropyStage1[i] = hb.EntropyStage1

	return 0
}

func (s *State) participantOrdering() (participantOrdering [common.QuorumSize]ParticipantIndex) {
	for i := range participantOrdering {
		participantOrdering[i] = ParticipantIndex(i)
	}

	// shuffle the list to produce a random host ordering by swapping each
	// element with a random element in front of it
	for i := range participantOrdering {
		newIndex, err := s.RandInt(i, common.QuorumSize)
		if err != nil {
			log.Fatalln(err)
		}
		tmp := participantOrdering[newIndex]
		participantOrdering[newIndex] = participantOrdering[i]
		participantOrdering[i] = tmp
	}

	return
}

// Takes all of the heartbeats and uses them to advance to the next state
func (s *State) Compile() {
	// arrive at a host ordering
	// create a list representing each host
	participantOrdering := s.participantOrdering()

	for _, participant := range participantOrdering {
		// process received heartbeats [switch these to processHeartbeat()]
		// skip if no host
		if s.Participants[participant] == nil {
			continue
		}

		if len(s.Heartbeats[participant]) != 1 {
			s.tossParticipant(participant)
			continue
		}

		for _, hb := range s.Heartbeats[participant] {
			s.processHeartbeat(hb, participant)

			// clear the heartbeat from s.Heartbeats
			s.Heartbeats[participant] = make(map[crypto.TruncatedHash]*Heartbeat)
		}
	}

	// move UpcomingEntropy to CurrentEntropy
	s.CurrentEntropy = s.UpcomingEntropy

	// generate a new heartbeat
	hb, err := s.NewHeartbeat()
	if err != nil {
		log.Fatalln(err)
	}
	signedHeartbeat, err := s.SignHeartbeat(hb)
	if err != nil {
		log.Fatalln(err)
	}

	// add our heartbeat to our heartbeat map
	s.Heartbeats[s.ParticipantIndex][signedHeartbeat.HeartbeatHash] = signedHeartbeat.Heartbeat

	// send the signed heartbeat to everyone
}

// Tick() updates s.CurrentStep, and calls Compile() when all steps are complete
// Tick() runs in its own gothread, only one instance of Tick() runs per state
func (s *State) Tick() {
	// check that no other instance of Tick() is running
	s.TickLock.Lock()
	if s.Ticking {
		s.TickLock.Unlock()
		return
	} else {
		s.Ticking = true
		s.TickLock.Unlock()
	}

	// Every common.StepDuration, advance the state stage
	ticker := time.Tick(common.StepDuration)
	for _ = range ticker {
		if s.CurrentStep == common.QuorumSize {
			s.Compile()
			s.CurrentStep = 1
		} else {
			s.CurrentStep += 1
		}
	}

	// if every we add code to stop ticking, this will be needed
	s.TickLock.Lock()
	s.Ticking = false
	s.TickLock.Unlock()
}

func (s *State) Start() {
	// send empty heartbeat to everybody

	go s.Tick()
}
