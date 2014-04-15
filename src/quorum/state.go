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
	// PreviousEntropy [common.QuorumSize]*crypto.TruncatedHash

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

func HandleMessage(m common.Message) {
	// take the payload and squeeze out the type bytes
	// use a switch statement based on type
}
