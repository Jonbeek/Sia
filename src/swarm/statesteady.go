package swarm

import (
	"common"
)

type StateSteady struct {
}

func (s *StateSteady) HandleTransaction(t common.Transaction) {
	return
}

func (s *StateSteady) HandleBlock(b *Block) State {
	return s
}

func NewStateSteady() State {
	return &StateSteady{}
}
