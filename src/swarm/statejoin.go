package swarm

import (
	"common"
)

func NewStateSwarmJoin(b *BlockChain) State {
	return &StateSwarmJoin{}
}

type StateSwarmJoin struct {
}

func (s *StateSwarmJoin) HandleTransaction(t common.Transaction) {
	return
}

func (s *StateSwarmJoin) HandleBlock(b *Block) State {
	return s
}
