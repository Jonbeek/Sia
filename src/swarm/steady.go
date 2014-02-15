package swarm

import (
	"common"
	"time"
)

type StateSteady struct {
	blockgen <-chan time.Time

	chain  *BlockChain
	block  *Block
	Stage2 map[string]string
}

func NewStateSteady(block *Block) (s *StateSteady) {
	s := make(StateSteady)
	s.block = block
	// Deal with the blockgen
	return
}

func (s *StateSteady) HandleTransaction(t common.Transaction) {
	switch n := t.(type) {
	case *HeartbeatTransaction:
		if VerifyHeartbeat(s.block, t) {
			s.Stage2[t.SwarmId()] = t.GetStage2()
		}
	default:
		return
	}
}

func (s *StateSteady) HandleBlock(b *Block) *Block {
	switch n := t.(type) {
	case *Block:
		// Check consistency with received Heartbeats
		s.block = b
		s.chain.AddBlock(b)
		return b
	default:
		return nil
	}
}
