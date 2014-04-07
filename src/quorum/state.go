package quorum

import (
	"common"
	"common/crypto"
)

type State struct {
	CurrentStep  int
	Participants []crypto.PublicKey

	Wallets map[string]uint64
}

func HandleMessage(m common.Message) {
	// take the payload and squeeze out the type bytes
	// use a switch statement based on type
}
