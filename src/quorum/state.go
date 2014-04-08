package quorum

import (
	"common"
	"common/crypto"
)

type State struct {
	// Who is participating in the quorum
	Participants map[crypto.PublicKey]*Participant

	// The cryptographic keys belonging to this host specifically
	PublicKey crypto.PublicKey
	SecretKey crypto.SecretKey

	// Consensus Algorithm Variables
	CurrentStep int
	Heartbeats  map[crypto.PublicKey]*Heartbeat

	// Wallet Data
	Wallets map[string]uint64
}

type Participant struct {
	Address string
}

func (p *Participant) IsEmpty() (rv bool) {
	if p.Address != "" {
		rv = false
		return
	}

	rv = true
	return
}

func HandleMessage(m common.Message) {
	// take the payload and squeeze out the type bytes
	// use a switch statement based on type
}
