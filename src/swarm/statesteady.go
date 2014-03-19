package swarm

import (
	"common"
	"crypto/sha256"
	"log"
	"time"
)

/* StateSteady

   This State runs an arbitrary swarm. It uses the three phase
   algorithm.

*/
type StateSteady struct {
	chain  *Blockchain
	stage2 string
	*ThreePhase
}

func NewStateSteady(chain *Blockchain, starttime time.Time, stage2 string) State {
	s := new(StateSteady)
	s.chain = chain
	s.stage2 = stage2
	s.ThreePhase = NewThreePhase(starttime, s)

	return s
}

func (s *StateSteady) Die() {
	s.ThreePhase.Die()
}

func (s *StateSteady) SendUpdate(u common.Update) {
	s.chain.outgoingUpdates <- u
}

func (s *StateSteady) ProduceHeartbeat() common.Update {

	stage1, stage2 := common.HashedRandomData(sha256.New(), 8)
	h := NewHeartbeat(s.chain.Id, s.chain.Host, stage1, s.stage2)
	s.stage2 = stage2
	return h
}

func (s *StateSteady) SignHeartbeat(h *Heartbeat) string {
	return h.Id
}

func (s *StateSteady) NewHostOk(hostid string) bool {
	// TODO: actually write this function
	return hostid == "joininghost"
}

func (s *StateSteady) ValidateHeartbeat(h *Heartbeat) bool {
	p, ok := s.chain.LastBlock().Heartbeats[h.Host]
	if !ok && s.NewHostOk(h.Host) {
		return true
	}
	if !ok {
		log.Print("STATESTEADY: previous beat not found")
		return false
	}
	ok = common.Hash(sha256.New(), h.EntropyStage2) == p.EntropyStage1
	if !ok {
		log.Print(p)
		log.Print(h)
		log.Print("STATESTEADY: hash not match")
		log.Print(common.Hash(sha256.New(), h.EntropyStage2))
		log.Print(p.EntropyStage1)
	}
	return ok
}

func (s *StateSteady) ValidateSignature(h *Heartbeat, sig string) bool {
	return sig == h.Id
}

func (s *StateSteady) HandleBlock(b *Block, _ time.Time) {
	s.chain.AddBlock(b)
}

func (s *StateSteady) Id() string {
	return s.chain.Id
}

func (s *StateSteady) Host() string {
	return s.chain.Host
}
