package swarm

import (
	"common"
)

func NewStateSwarmConnected() State {
	return &StateSwarmConnected{}
}

type StateSwarmConnected struct {
}

func (s *StateSwarmConnected) HandleTransaction(t common.Transaction) {
	return
}

func (s *StateSwarmConnected) HandleBlock(b *Block) State {
	return s
}
