package quorum

import (
	"common"
	"common/crypto"
)

// The state is what provides persistence to the consensus algorithms.
// The state is updated every block, and every honest host is
// guaranteed to update their state in the same way.
type State struct {
	// Who is participating in the quorum
	Participants map[crypto.PublicKey]*Participant

	// The cryptographic keys belonging to this host specifically
	PublicKey crypto.PublicKey
	SecretKey crypto.SecretKey

	// Consensus Algorithm Variables
	CurrentStep int
	Heartbeats  map[crypto.PublicKey]map[crypto.Hash]*Heartbeat

	// Wallet Data
	Wallets map[string]uint64
}

// Currently just an address, as the participant is accessed
// by knowing the public key. It's in its own struct because
// more fields might be added.
type Participant struct {
	Address string
}

func CreateState() (s State, err error) {
	// initialize crypto keys
	pubKey, secKey, err := crypto.CreateKeyPair()
	if err != nil {
		// some error
	}
	s.PublicKey = pubKey
	s.SecretKey = secKey

	s.CurrentStep = 1
	s.Heartbeats = make(map[crypto.PublicKey]map[crypto.Hash]*Heartbeat)
	s.Wallets = make(map[string]uint64)
	return
}

func HandleMessage(m common.Message) {
	// take the payload and squeeze out the type bytes
	// use a switch statement based on type
}
