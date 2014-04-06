package quorum

import (
	"common"
	"common/log"
	"sync"
	"time"
)

/* A slightly generic implementation of the three phase concensus algorithm

   The first phase is a pause until the designated swarm start time, this allows
   time for hosts to come online.

   Second each host sends a heartbeat to all other hosts

   Third each hosts signs all heartbeats they have seen and sends them to all
   other hosts

   Fourth each hosts compiles a block which includes all heartbeats which have
   been signed by 51% of the network

   The final block is the combination of all heartbeat transactions which have
   been signed by 51% of the network

*/

type NetHandler interface {
	SendUpdate(u common.Update)
	ProduceHeartbeat() common.Update
	ValidateHeartbeat(h *Heartbeat) bool
	SignHeartbeat(h *Heartbeat) string
	ValidateSignature(h *Heartbeat, sig string) bool
	Id() string
	Host() string
	HandleBlock(b *Block, nextphasetime time.Time)
}

type ThreePhase struct {
	lock          *sync.Mutex
	heartbeats    map[string]*Heartbeat
	signatures    map[string]map[string]string
	nextphase     string
	phasetimer    <-chan time.Time
	nextphasetime time.Time
	handler       NetHandler
}

func NewThreePhase(starttime time.Time, handler NetHandler) (s *ThreePhase) {

	s = new(ThreePhase)
	s.phasetimer = time.After(starttime.Sub(time.Now()))
	s.nextphase = "Heartbeat"
	s.nextphasetime = starttime

	s.lock = &sync.Mutex{}
	s.heartbeats = make(map[string]*Heartbeat)
	s.signatures = make(map[string]map[string]string)
	s.handler = handler

	go s.mainloop()
	return
}

func (s *ThreePhase) Die() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.nextphase = "Dead"
}

func (s *ThreePhase) mainloop() {

	for _ = range s.phasetimer {
		log.Debug("State: phasetimer fired")
		s.lock.Lock()
		s.nextphasetime = s.nextphasetime.Add(common.STATEINFORMEDDELTA)
		switch s.nextphase {

		case "Heartbeat":
			log.Debug("3PHASE: Heartbeat phase")
			s.phasetimer = time.Tick(common.STATEINFORMEDDELTA)
			h := s.handler.ProduceHeartbeat()
			go s.handler.SendUpdate(h)
			s.nextphase = "HeartbeatSigning"
			s.lock.Unlock()
			defer s.mainloop()
			return

		case "HeartbeatSigning":
			log.Debug("3PHASE: HeartbeatSigning phase")
			s.phasetimer = time.Tick(common.STATEINFORMEDDELTA)
			s.produceSignedHeartbeats()
			s.nextphase = "BlockGeneration"

		case "BlockGeneration":
			log.Debug("3PHASE: BlockGeneration phase")
			s.produceBlock()
			s.nextphase = "BlockDecision"

		case "BlockDecision":
			log.Debug("3PHASE: BlockDecision phase")
			ok := s.decideBlock()
			log.Debug("3PHASE: ok=", ok)
			s.nextphase = "Heartbeat"
			s.heartbeats = make(map[string]*Heartbeat)
			s.signatures = make(map[string]map[string]string)

		case "Dead":
			defer s.lock.Unlock()
			return

		default:
			panic("Unrecognized state: " + s.nextphase)
		}
		s.lock.Unlock()
	}
}

func (s *ThreePhase) produceSignedHeartbeats() {

	signatures := make(map[string]string)
	for id, h := range s.heartbeats {
		if !s.handler.ValidateHeartbeat(h) {
			log.Debug("3PHASE: Validate Heartbeat Failed")
			continue
		}

		signatures[id] = s.handler.SignHeartbeat(h)
	}

	h := NewHeartbeatList(s.handler.Id(), s.handler.Host(), s.heartbeats, signatures)
	go s.handler.SendUpdate(h)
}

func (s *ThreePhase) copyMaps() (map[string]*Heartbeat, map[string]map[string]string) {
	heartbeats := make(map[string]*Heartbeat)
	signatures := make(map[string]map[string]string)

	for id, heartbeat := range s.heartbeats {
		heartbeats[id] = heartbeat
		if signatures[id] == nil {
			signatures[id] = make(map[string]string)
		}
		for host, sign := range s.signatures[id] {
			signatures[id][host] = sign
		}
	}
	return heartbeats, signatures
}

func (s *ThreePhase) produceBlock() {

	heartbeats, signatures := s.copyMaps()
	b := NewBlock(s.handler.Id(), s.handler.Host(), heartbeats, signatures, s.nextphasetime.Add(common.STATEINFORMEDDELTA))
	go s.handler.SendUpdate(b)
}

func (s *ThreePhase) decideBlock() bool {
	log.Debug("3PHASE: decideBlock")
	heartbeats, signatures := s.copyMaps()

	for id, _ := range heartbeats {
		log.Debug("Heartbeat id=", id, " len(signatures)=", len(signatures[id]))
		if len(signatures[id]) <= common.SWARMSIZE/2 {
			delete(heartbeats, id)
			delete(signatures, id)
		}
	}

	if len(heartbeats) <= common.SWARMSIZE/2 {
		return false
	}

	hostheartbeats := make(map[string]*Heartbeat)
	for _, hb := range heartbeats {
		hostheartbeats[hb.Host] = hb
	}

	b := NewBlock(s.handler.Id(), s.handler.Host(), hostheartbeats,
		signatures, s.nextphasetime)
	s.handler.HandleBlock(b, s.nextphasetime)

	return true

}

func (s *ThreePhase) HandleUpdate(t common.Update) {

	s.lock.Lock()
	defer s.lock.Unlock()

	switch n := t.(type) {
	case *Heartbeat:
		if s.nextphase != "HeartbeatSigning" &&
			s.nextphase != "Heartbeat" {
			return
		}

		s.heartbeats[n.Id] = n
		log.Debug("3PHASE: Node added")

	case *HeartbeatList:
		if s.nextphase != "BlockGeneration" &&
			s.nextphase != "HeartbeatSigning" {
			return
		}

		for id, signature := range n.Signatures {
			if !s.handler.ValidateSignature(n.Heartbeats[id], signature) {
				continue
			}
			if _, ok := s.heartbeats[id]; !ok {
				s.heartbeats[id] = n.Heartbeats[id]
			}
			if _, ok := s.signatures[id]; !ok {
				s.signatures[id] = make(map[string]string)
			}
			s.signatures[id][n.Host] = signature
		}

	case *Block:
		if s.nextphase != "BlockDecision" &&
			s.nextphase != "BlockGeneration" {
			return
		}

		for id, signatures := range n.Signatures {

			if _, ok := s.heartbeats[id]; !ok {
				s.heartbeats[id] = n.Heartbeats[id]
			}
			if _, ok := s.signatures[id]; !ok {
				s.signatures[id] = make(map[string]string)
			}
			for host, signature := range signatures {
				if !s.handler.ValidateSignature(n.Heartbeats[id], signature) {
					continue
				}
				s.signatures[id][host] = signature
			}
		}

	default:

	}
	return
}
