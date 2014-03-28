package swarm

import (
	"common"
	"crypto/sha256"
	"time"
)

/* SwarmInformed

   This State initializes the arbitrary swarm. It uses the three phase
   algorithm.

*/
type StateSwarmInformed struct {
	chain  *Blockchain
	stage2 string
	*ThreePhase
}

func NewStateSwarmInformed(chain *Blockchain, starttime time.Time) State {
	s := new(StateSwarmInformed)
	s.chain = chain
	s.ThreePhase = NewThreePhase(starttime, s)

	return s
}

func (s *StateSwarmInformed) Die() {
	s.ThreePhase.Die()
}

func (s *StateSwarmInformed) SendUpdate(u common.Update) {
	s.chain.outgoingUpdates <- u
}

func (s *StateSwarmInformed) ProduceHeartbeat() common.Update {

	stage1, stage2 := common.HashedRandomData(sha256.New(), 8)
	s.stage2 = stage2

	h := NewHeartbeat(s.chain.Id, s.chain.Host, stage1, "", nil)
	return h
}

func (s *StateSwarmInformed) SignHeartbeat(h *Heartbeat) string {
	return h.Id
}

func (s *StateSwarmInformed) ValidateHeartbeat(h *Heartbeat) bool {
	return len(h.EntropyStage1) > 0
}

func (s *StateSwarmInformed) ValidateSignature(h *Heartbeat, sig string) bool {
	return sig == h.Id
}

func (s *StateSwarmInformed) HandleBlock(b *Block, next time.Time) {
	s.chain.AddBlock(b)
	s.chain.SwitchState(NewStateSteady(s.chain, next, s.stage2))
	go s.Die()
}

func (s *StateSwarmInformed) Id() string {
	return s.chain.Id
}

func (s *StateSwarmInformed) Host() string {
	return s.chain.Host
}
