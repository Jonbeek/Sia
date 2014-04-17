package quorum

import (
	"common"
	"common/crypto"
	"fmt"
)

type ParticipantIndex uint8

// The state is what provides persistence to the consensus algorithms.
// The state is updated every block, and every honest host is
// guaranteed to update their state in the same way.
type State struct {
	// Who is participating in the quorum
	Participants [common.QuorumSize]*Participant
	// Our own index in the quorum
	ParticipantIndex ParticipantIndex

	// The cryptographic keys belonging to this host specifically
	PublicKey crypto.PublicKey
	SecretKey crypto.SecretKey

	// The hash of this was used in stage 1 for the most recent heartbeat
	StoredEntropyStage2 common.Entropy

	// The stage 1 entropies from the last block
	PreviousEntropyStage1 [common.QuorumSize]crypto.TruncatedHash
	// Entropy seed to be used while compiling next block
	CurrentEntropy common.Entropy
	// Entropy that gets built out as the block is compiled
	UpcomingEntropy common.Entropy

	// Consensus Algorithm Status
	CurrentStep int
	Heartbeats  [common.QuorumSize]map[crypto.TruncatedHash]*Heartbeat

	// Wallet Data
	Wallets map[string]uint64
}

// Currently just an address, as the participant is accessed
// by knowing the public key. It's in its own struct because
// more fields might be added.
type Participant struct {
	PublicKey crypto.PublicKey
}

// Create and initialize a state object
func CreateState(participantIndex ParticipantIndex) (s State, err error) {
	// check that participantIndex is legal
	if int(participantIndex) >= common.QuorumSize {
		err = fmt.Errorf("Invalid participant index!")
		return
	}

	// initialize crypto keys
	pubKey, secKey, err := crypto.CreateKeyPair()
	if err != nil {
		return
	}
	s.PublicKey = pubKey
	s.SecretKey = secKey

	s.ParticipantIndex = participantIndex

	s.CurrentStep = 1
	s.Wallets = make(map[string]uint64)

	// add ourselves to list of participants
	s.AddParticipant(s.PublicKey, participantIndex)

	// set the stored EntropyStage1 to be the hash of the zero value
	emptyHash, err := crypto.CalculateTruncatedHash(s.StoredEntropyStage2[:])
	if err != nil {
		return
	}
	for i := range s.PreviousEntropyStage1 {
		s.PreviousEntropyStage1[i] = emptyHash
	}

	// create our first heartbeat and add it to our heartbeat map
	hb, err := s.NewHeartbeat()
	if err != nil {
		return
	}

	// get the heartbeats hash
	heartbeatHash, err := crypto.CalculateTruncatedHash([]byte(hb.Marshal()))
	s.Heartbeats[participantIndex][heartbeatHash] = hb

	return
}

// Populates a state with this participant, initializing variables as needed
// return codes are arbitraily chosen and are only for the test suite
func (s *State) AddParticipant(pubKey crypto.PublicKey, i ParticipantIndex) (err error) {
	// Check that there is not already a participant for the index
	if s.Participants[i] != nil {
		err = fmt.Errorf("A participant already exists for the given index!")
		return
	}

	// initialize participant object
	var p Participant
	p.PublicKey = pubKey

	// initialize the heartbeat map for this participant
	s.Heartbeats[i] = make(map[crypto.TruncatedHash]*Heartbeat)

	// add to state
	s.Participants[i] = &p

	return
}

// Use the entropy stored in the state to generate a random
// integer [low, high)
func (s *State) RandInt(low int, high int) (randInt int, err error) {
	// verify there's a gap between the numbers
	if low == high {
		err = fmt.Errorf("low and high cannot be the same number")
		return
	}

	// Convert CurrentEntropy into an int
	rollingInt := 0
	for i := 0; i < 4; i++ {
		rollingInt = rollingInt << 4
		rollingInt += int(s.CurrentEntropy[0])
	}

	randInt = (rollingInt % (high - low)) + low

	// Convert random number seed to next value
	truncatedHash, err := crypto.CalculateTruncatedHash(s.CurrentEntropy[:])
	s.CurrentEntropy = common.Entropy(truncatedHash)
	return
}

func HandleMessage(m common.Message) {
	// take the payload and squeeze out the type bytes
	// use a switch statement based on type
}
