package quorum

import (
	"common"
	"common/crypto"
	"common/log"
	"fmt"
	"time"
)

// All information that needs to be passed between participants each block
type Heartbeat struct {
	EntropyStage1 crypto.TruncatedHash
	EntropyStage2 common.Entropy
}

// Contains a heartbeat that has been signed iteratively, is a key part of the
// signed solution to the Byzantine Generals Problem
type SignedHeartbeat struct {
	Heartbeat     *Heartbeat
	HeartbeatHash crypto.TruncatedHash
	Signatures    []crypto.Signature
	Signatories   []ParticipantIndex
}

// Using the current State, NewHeartbeat creates a heartbeat that fulfills all
// of the requirements of the quorum.
func (s *State) NewHeartbeat() (hb *Heartbeat, err error) {
	hb = new(Heartbeat)

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

func MarshalledHeartbeatLen() int {
	return crypto.TruncatedHashSize + common.EntropyVolume
}

// Convert Heartbeat to string
func (hb *Heartbeat) Marshal() (marshalledHeartbeat []byte) {
	marshalledHeartbeat = append(hb.EntropyStage1[:], hb.EntropyStage2[:]...)
	return
}

// Convert string to Heartbeat
func UnmarshalHeartbeat(marshalledHeartbeat []byte) (hb *Heartbeat, err error) {
	expectedLen := MarshalledHeartbeatLen()
	if len(marshalledHeartbeat) != expectedLen {
		err = fmt.Errorf("Marshalled heartbeat is the wrong size!")
		return
	}

	hb = new(Heartbeat)
	copy(hb.EntropyStage1[:], marshalledHeartbeat)
	copy(hb.EntropyStage2[:], marshalledHeartbeat[crypto.TruncatedHashSize:])
	return
}

func (s *State) SignHeartbeat(hb *Heartbeat) (sh *SignedHeartbeat, err error) {
	sh = new(SignedHeartbeat)

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

func (sh *SignedHeartbeat) Marshal() (msh []byte, err error) {
	// error check the input
	if len(sh.Signatures) > common.QuorumSize {
		err = fmt.Errorf("Too many signatures on heartbeat")
		return
	} else if len(sh.Signatures) != len(sh.Signatories) {
		err = fmt.Errorf("Mismatched set of signatures and signatories")
		return
	}

	// get all pieces of the marshalledSignedHeartbeat
	mhb := sh.Heartbeat.Marshal()
	numSignatures := byte(len(sh.Signatures))
	numBytes := len(mhb) + 1 + int(numSignatures)*(crypto.SignatureSize+1)
	msh = make([]byte, numBytes)

	index := 0
	copy(msh[index:], mhb)
	index += len(mhb)
	copy(msh[index:], string(numSignatures))
	index += 1
	for i := 0; i < int(numSignatures); i++ {
		copy(msh[index:], sh.Signatures[i][:])
		index += crypto.SignatureSize
		copy(msh[index:], string(sh.Signatories[i]))
		index += 1
	}

	return
}

func UnmarshalSignedHeartbeat(msh []byte) (sh *SignedHeartbeat, err error) {
	// error check the input
	if len(msh) <= MarshalledHeartbeatLen() {
		err = fmt.Errorf("input for UnmarshalSignedHeartbeat is too short")
		return
	}
	numSignatures := int(msh[MarshalledHeartbeatLen()])
	signatureSectionLen := numSignatures * (crypto.SignatureSize + 1)
	totalLen := MarshalledHeartbeatLen() + 1 + signatureSectionLen
	if len(msh) != totalLen {
		err = fmt.Errorf("input for UnmarshalSignedHeartbeat is incorrect length, expecting ", totalLen, " bytes")
		return
	}

	// get sh.Heartbeat and sh.HeartbeatHash
	sh = new(SignedHeartbeat)
	index := 0
	heartbeat, err := UnmarshalHeartbeat(msh[index:MarshalledHeartbeatLen()])
	if err != nil {
		return
	}
	heartbeatHash, err := crypto.CalculateTruncatedHash(msh[index:MarshalledHeartbeatLen()])
	if err != nil {
		return
	}
	sh.Heartbeat = heartbeat
	sh.HeartbeatHash = heartbeatHash

	// get sh.Signatures and sh.Signatories
	index += MarshalledHeartbeatLen()
	index += 1
	sh.Signatures = make([]crypto.Signature, numSignatures)
	sh.Signatories = make([]ParticipantIndex, numSignatures)
	for i := 0; i < numSignatures; i++ {
		copy(sh.Signatures[i][:], msh[index:])
		index += crypto.SignatureSize
		sh.Signatories[i] = ParticipantIndex(msh[index])
		index += 1
	}

	return
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

func (s *State) announceSignedHeartbeat(sh *SignedHeartbeat) {
	for i := range s.Participants {
		if s.Participants[i] != nil {
			payload, err := sh.Marshal()
			if err != nil {
				log.Fatalln(err)
			}

			m := new(common.Message)
			m.Payload = append([]byte{byte(1)}, payload...)
			m.Destination = s.Participants[i].Address
			err = s.MessageSender.SendMessage(m)
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
func (s *State) HandleSignedHeartbeat(message []byte) (returnCode int) {
	print(s.ParticipantIndex)
	println(" received message: ")
	// covert message to SignedHeartbeat
	sh, err := UnmarshalSignedHeartbeat(message)
	if err != nil {
		log.Infoln("Received bad message SignedHeartbeat: ", err)
		return
	}

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
	_, err = crypto.Sign(s.SecretKey, signedMessage.Message)
	if err != nil {
		log.Fatalln(err)
	}

	// Send the new message to everybody

	returnCode = 0
	return
}

// participants are processed in a random order each block, determied by the
// entropy for the block
func (s *State) participantOrdering() (participantOrdering [common.QuorumSize]ParticipantIndex) {
	// create an in-order list of participants
	for i := range participantOrdering {
		participantOrdering[i] = ParticipantIndex(i)
	}

	// shuffle the list of participants
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

// Removes all traces of a participant from the State
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

// Takes all of the heartbeats and uses them to advance to the next state
func (s *State) Compile() {
	participantOrdering := s.participantOrdering()

	for _, participant := range participantOrdering {
		if s.Participants[participant] == nil {
			continue
		}

		// each participant must submit exactly 1 heartbeat
		if len(s.Heartbeats[participant]) != 1 {
			s.tossParticipant(participant)
			continue
		}

		for _, hb := range s.Heartbeats[participant] {
			s.processHeartbeat(hb, participant)
		}
		// clear map of heartbeats for next block
		s.Heartbeats[participant] = make(map[crypto.TruncatedHash]*Heartbeat)
	}

	// move UpcomingEntropy to CurrentEntropy
	s.CurrentEntropy = s.UpcomingEntropy

	// generate a new heartbeat, add it to our map, and announce it
	hb, err := s.NewHeartbeat()
	if err != nil {
		log.Fatalln(err)
	}
	hash, err := crypto.CalculateTruncatedHash(hb.Marshal())
	if err != nil {
		log.Fatalln(err)
	}

	s.Heartbeats[s.ParticipantIndex][hash] = hb
	shb, err := s.SignHeartbeat(hb)
	if err != nil {
		log.Fatalln(err)
	}
	s.announceSignedHeartbeat(shb)
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
