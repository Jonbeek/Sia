package swarm

import (
	"common"
	"crypto/sha256"
	"time"
)

/* StateJoin

   This State runs an arbitrary swarm. It uses the three phase
   algorithm.

*/
type StateJoin struct {
	chain  *Blockchain
	stage2 string
	*ThreePhase
}

func NewStateJoin(chain *Blockchain, starttime time.Time) State {
	s := new(StateJoin)
	s.chain = chain
	s.ThreePhase = NewThreePhase(starttime, s)

	return s
}

func (s *StateJoin) Die() {
	s.ThreePhase.Die()
}

func (s *StateJoin) SendUpdate(u common.Update) {
	s.chain.outgoingUpdates <- u
}

func (s *StateJoin) ProduceHeartbeat() common.Update {

	stage1, stage2 := common.HashedRandomData(sha256.New(), 8)
	h := NewHeartbeat(s.chain.Id, s.chain.Host, stage1, "")
	s.stage2 = stage2
	return h
}

func (s *StateJoin) SignHeartbeat(h *Heartbeat) string {
	return h.Id
}

func (s *StateJoin) ValidateHeartbeat(h *Heartbeat) bool {
	return true
}

func (s *StateJoin) ValidateSignature(h *Heartbeat, sig string) bool {
	return sig == h.Id
}

func (s *StateJoin) HandleBlock(b *Block, next time.Time) {
	s.chain.AddBlock(b)
	if s.chain.HostActive(s.Host()) {
		s.chain.SwitchState(NewStateSteady(s.chain, next, s.stage2))
		go s.Die()
	}
}

func (s *StateJoin) Id() string {
	return s.chain.Id
}

func (s *StateJoin) Host() string {
	return s.chain.Host
}
