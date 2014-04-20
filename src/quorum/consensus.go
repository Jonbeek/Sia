package quorum

import (
	"common"
	"common/crypto"
	"common/log"
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
type signedHeartbeat struct {
	heartbeat     *heartbeat
	heartbeatHash crypto.TruncatedHash
	signatures    []crypto.Signature
	signatories   []participantIndex
}

// Using the current State, newHeartbeat() creates a heartbeat that fulfills all
// of the requirements of the quorum.
func (s *State) newHeartbeat() (hb *heartbeat, err error) {
	hb = new(heartbeat)

	// Fetch value used to produce EntropyStage1 in prev. heartbeat
	hb.entropyStage2 = s.storedEntropyStage2

	// Generate EntropyStage2 for next heartbeat
	rawEntropy, err := crypto.RandomByteSlice(common.EntropyVolume)
	if err != nil {
		return
	}
	copy(s.storedEntropyStage2[:], rawEntropy)

	// Use EntropyStage2 to generate EntropyStage1 for this heartbeat
	hb.entropyStage1, err = crypto.CalculateTruncatedHash(s.storedEntropyStage2[:])
	if err != nil {
		return
	}

	// more code will be added here

	return
}

func marshalledHeartbeatLen() int {
	return crypto.TruncatedHashSize + common.EntropyVolume
}

// Convert Heartbeat to string
func (hb *heartbeat) marshal() (marshalledHeartbeat []byte) {
	marshalledHeartbeat = append(hb.entropyStage1[:], hb.entropyStage2[:]...)
	return
}

// Convert string to Heartbeat
func unmarshalHeartbeat(marshalledHeartbeat []byte) (hb *heartbeat, err error) {
	expectedLen := marshalledHeartbeatLen()
	if len(marshalledHeartbeat) != expectedLen {
		err = fmt.Errorf("Marshalled heartbeat is the wrong size!")
		return
	}

	hb = new(heartbeat)
	copy(hb.entropyStage1[:], marshalledHeartbeat)
	copy(hb.entropyStage2[:], marshalledHeartbeat[crypto.TruncatedHashSize:])
	return
}

func (s *State) signHeartbeat(hb *heartbeat) (sh *signedHeartbeat, err error) {
	sh = new(signedHeartbeat)

	// confirm heartbeat and hash
	sh.heartbeat = hb
	marshalledHb := hb.marshal()
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash([]byte(marshalledHb))
	if err != nil {
		return
	}

	// fill out sigantures
	sh.signatures = make([]crypto.Signature, 1)
	signedHb, err := crypto.Sign(s.secretKey, string(sh.heartbeatHash[:]))
	if err != nil {
		return
	}
	sh.signatures[0] = signedHb.Signature
	sh.signatories = make([]participantIndex, 1)
	sh.signatories[0] = s.participantIndex
	return
}

func (sh *signedHeartbeat) marshal() (msh []byte, err error) {
	// error check the input
	if len(sh.signatures) > common.QuorumSize {
		err = fmt.Errorf("Too many signatures on heartbeat")
		return
	} else if len(sh.signatures) != len(sh.signatories) {
		err = fmt.Errorf("Mismatched set of signatures and signatories")
		return
	}

	// get all pieces of the marshalledSignedHeartbeat
	mhb := sh.heartbeat.marshal()
	numSignatures := byte(len(sh.signatures))
	numBytes := len(mhb) + 1 + int(numSignatures)*(crypto.SignatureSize+1)
	msh = make([]byte, numBytes)

	index := 0
	copy(msh[index:], mhb)
	index += len(mhb)
	copy(msh[index:], string(numSignatures))
	index += 1
	for i := 0; i < int(numSignatures); i++ {
		copy(msh[index:], sh.signatures[i][:])
		index += crypto.SignatureSize
		copy(msh[index:], string(sh.signatories[i]))
		index += 1
	}

	return
}

func unmarshalSignedHeartbeat(msh []byte) (sh *signedHeartbeat, err error) {
	// error check the input
	if len(msh) <= marshalledHeartbeatLen() {
		err = fmt.Errorf("input for unmarshalSignedHeartbeat is too short")
		return
	}
	numSignatures := int(msh[marshalledHeartbeatLen()])
	signatureSectionLen := numSignatures * (crypto.SignatureSize + 1)
	totalLen := marshalledHeartbeatLen() + 1 + signatureSectionLen
	if len(msh) != totalLen {
		err = fmt.Errorf("input for UnmarshalSignedHeartbeat is incorrect length, expecting ", totalLen, " bytes")
		return
	}

	// get sh.Heartbeat and sh.HeartbeatHash
	sh = new(signedHeartbeat)
	index := 0
	heartbeat, err := unmarshalHeartbeat(msh[index:marshalledHeartbeatLen()])
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash(msh[index:marshalledHeartbeatLen()])
	if err != nil {
		return
	}
	sh.heartbeat = heartbeat
	sh.heartbeatHash = heartbeatHash

	// get sh.Signatures and sh.Signatories
	index += marshalledHeartbeatLen()
	index += 1
	sh.signatures = make([]crypto.Signature, numSignatures)
	sh.signatories = make([]participantIndex, numSignatures)
	for i := 0; i < numSignatures; i++ {
		copy(sh.signatures[i][:], msh[index:])
		index += crypto.SignatureSize
		sh.signatories[i] = participantIndex(msh[index])
		index += 1
	}

	return
}

func (s *State) processHeartbeat(hb *heartbeat, i participantIndex) int {
	// compare EntropyStage2 to the hash from the previous heartbeat
	expectedHash, err := crypto.CalculateTruncatedHash(hb.entropyStage2[:])
	if err != nil {
		log.Fatalln(err)
	}
	if expectedHash != s.previousEntropyStage1[i] {
		s.tossParticipant(i)
		return 1
	}

	// Add the EntropyStage2 to UpcomingEntropy
	th, err := crypto.CalculateTruncatedHash(append(s.upcomingEntropy[:], hb.entropyStage2[:]...))
	s.upcomingEntropy = common.Entropy(th)
	// update PreviousEntropy, to compare this EntropyStage1 against the next
	// EntropyStage1
	s.previousEntropyStage1[i] = hb.entropyStage1

	return 0
}

func (s *State) announceSignedHeartbeat(sh *signedHeartbeat) {
	for i := range s.participants {
		if s.participants[i] != nil {
			payload, err := sh.marshal()
			if err != nil {
				log.Fatalln(err)
			}

			m := new(common.Message)
			m.Payload = append([]byte{byte(1)}, payload...)
			m.Destination = s.participants[i].Address
			//time.Sleep(time.Millisecond) // prevents panics. No idea where original source of bug is.
			err = s.messageSender.SendMessage(m)
			if err != nil {
				log.Fatalln("Error while sending message")
			}
		}
	}
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
func (s *State) handleSignedHeartbeat(message []byte) (returnCode int) {
	// covert message to SignedHeartbeat
	sh, err := unmarshalSignedHeartbeat(message)
	if err != nil {
		log.Infoln("Received bad message SignedHeartbeat: ", err)
		return
	}

	// Check that the slices of signatures and signatories are of the same length
	if len(sh.signatures) != len(sh.signatories) {
		log.Infoln("SignedHeartbeat has mismatched signatures")
		returnCode = 1
		return
	}

	// s.CurrentStep must be less than or equal to len(sh.Signatories), unless the
	// current step is common.QuorumSize and len(sh.Signatories) == 1
	if s.currentStep > len(sh.signatories) {
		if s.currentStep == common.QuorumSize && len(sh.signatories) == 1 {
			// sleep long enough to pass the first requirement
			time.Sleep(common.StepDuration)
			// now continue to rest of function
		} else {
			log.Infoln("Received an out-of-sync SignedHeartbeat")
			returnCode = 2
			return
		}
	}

	// Check bounds on first signatory
	if int(sh.signatories[0]) >= common.QuorumSize {
		log.Infoln("Received an out of bounds index")
		returnCode = 9
		return
	}

	// Check existence of first signatory
	if s.participants[sh.signatories[0]] == nil {
		log.Infoln("Received heartbeat from non-participant")
		returnCode = 10
		return
	}

	// Check if we have already received this heartbeat
	_, exists := s.heartbeats[sh.signatories[0]][sh.heartbeatHash]
	if exists {
		returnCode = 8
		return
	}

	// while processing signatures, signedMessage will keep growing
	var signedMessage crypto.SignedMessage
	signedMessage.Message = string(sh.heartbeatHash[:])
	// keep a map of which signatories have already been confirmed
	previousSignatories := make(map[participantIndex]bool)

	for i, signatory := range sh.signatories {
		// Check bounds on the signatory
		if int(signatory) >= common.QuorumSize {
			log.Infoln("Received an out of bounds index")
			returnCode = 9
			return
		}

		// Verify that the signatory is a participant in the quorum
		if s.participants[signatory] == nil {
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
		signedMessage.Signature = sh.signatures[i]
		verification, err := crypto.Verify(s.participants[signatory].PublicKey, signedMessage)
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
	// Don't check if heartbeat is valid, that's for compile()
	s.heartbeats[sh.signatories[0]][sh.heartbeatHash] = sh.heartbeat

	// Sign the stack of signatures and send it to all hosts
	_, err = crypto.Sign(s.secretKey, signedMessage.Message)
	if err != nil {
		log.Fatalln(err)
	}

	// Send the new message to everybody

	returnCode = 0
	return
}

// participants are processed in a random order each block, determied by the
// entropy for the block. participantOrdering() deterministically picks that
// order, using entropy from the state.
func (s *State) participantOrdering() (participantOrdering [common.QuorumSize]participantIndex) {
	// create an in-order list of participants
	for i := range participantOrdering {
		participantOrdering[i] = participantIndex(i)
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
func (s *State) tossParticipant(pi participantIndex) {
	// remove from s.Participants
	s.participants[pi] = nil

	// remove from s.PreviousEntropyStage1
	var emptyEntropy common.Entropy
	zeroHash, err := crypto.CalculateTruncatedHash(emptyEntropy[:])
	s.previousEntropyStage1[pi] = zeroHash
	if err != nil {
		log.Fatal(err)
	}

	// nil map in s.Heartbeats
	s.heartbeats[pi] = nil
}

// compile() takes the list of heartbeats and uses them to advance the state.
func (s *State) compile() {
	participantOrdering := s.participantOrdering()

	// Read read heartbeats, process them, then archive them. Other functions
	// concurrently access the heartbeats, so mutexes are needed.
	for _, participant := range participantOrdering {
		if s.participants[participant] == nil {
			continue
		}

		// each participant must submit exactly 1 heartbeat
		if len(s.heartbeats[participant]) != 1 {
			s.tossParticipant(participant)
			continue
		}

		for _, hb := range s.heartbeats[participant] {
			s.processHeartbeat(hb, participant)
		}

		// archive heartbeats
		// currently, archives are sent to /dev/null
		s.heartbeats[participant] = make(map[crypto.TruncatedHash]*heartbeat)
	}

	// move UpcomingEntropy to CurrentEntropy
	s.currentEntropy = s.upcomingEntropy

	// generate a new heartbeat and add it to s.Heartbeats
	hb, err := s.newHeartbeat()
	if err != nil {
		log.Fatalln(err)
	}
	hash, err := crypto.CalculateTruncatedHash(hb.marshal())
	if err != nil {
		log.Fatalln(err)
	}
	s.heartbeats[s.participantIndex][hash] = hb

	// sign and annouce the heartbeat
	shb, err := s.signHeartbeat(hb)
	if err != nil {
		log.Fatalln(err)
	}
	s.announceSignedHeartbeat(shb)
}

// Tick() updates s.CurrentStep, and calls compile() when all steps are complete
// Tick() runs in its own gothread, only one instance of Tick() runs per state
func (s *State) tick() {
	// check that no other instance of Tick() is running
	s.tickLock.Lock()
	if s.ticking {
		s.tickLock.Unlock()
		return
	} else {
		s.ticking = true
		s.tickLock.Unlock()
	}

	// Every common.StepDuration, advance the state stage
	ticker := time.Tick(common.StepDuration)
	for _ = range ticker {
		s.lock.Lock()
		if s.currentStep == common.QuorumSize {
			s.compile()
			s.currentStep = 1
		} else {
			s.currentStep += 1
		}
		s.lock.Unlock()
	}

	// if every we add code to stop ticking, this will be needed
	s.tickLock.Lock()
	s.ticking = false
	s.tickLock.Unlock()
}
