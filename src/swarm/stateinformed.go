package swarm

import (
	"common"
	"crypto/sha256"
	"log"
	"sync"
	"time"
)

/* SwarmInformed

   This State initializes the arbitrary swarm. It uses a simple consensus
   algorithm.

   The first phase is a pause until the designated swarm start time, this allows
   time for hosts to come online.

   Second each host sends a heartbeat to all ovther hosts

   Third each hosts signs all heartbeats they have seen and sends them to all
   other hosts

   Fourth each hosts compiles a block which includes all heartbeats which have
   been signed by 51% of the network

   The final block is the combination of all heartbeat transactions which have
   been signed by 51% of the network

*/
type StateSwarmInformed struct {
	chain      *Blockchain
	lock       *sync.Mutex
	heartbeats map[string]*Heartbeat
	signatures map[string]map[string]string
	stage2     string
	nextphase  string
	phasetimer <-chan time.Time
}

func NewStateSwarmInformed(chain *Blockchain, starttime time.Time) (s *StateSwarmInformed) {
	s = new(StateSwarmInformed)
	s.chain = chain

	// Wait until the swarm is supposed to start before producing a heartbeat
	s.phasetimer = time.After(starttime.Sub(time.Now()))
	s.nextphase = "Heartbeat"

	s.lock = &sync.Mutex{}
	s.heartbeats = make(map[string]*Heartbeat)
	s.signatures = make(map[string]map[string]string)

	go s.mainloop()
	return
}

func (s *StateSwarmInformed) Die() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.nextphase = "Dead"
}

func (s *StateSwarmInformed) sendUpdate(u common.Update) {
	s.chain.outgoingUpdates <- u
}

func (s *StateSwarmInformed) mainloop() {

	for _ = range s.phasetimer {
		log.Print("State: phasetimer fired")
		s.lock.Lock()
		switch s.nextphase {

		case "Heartbeat":
			log.Print("STATE: Heartbeat phase")
			s.phasetimer = time.Tick(common.STATEINFORMEDDELTA)
			s.produceHeartbeat()
			s.nextphase = "HeartbeatSigning"
			s.lock.Unlock()
			s.mainloop()
			return

		case "HeartbeatSigning":
			log.Print("STATE: HeartbeatSigning phase")
			s.phasetimer = time.Tick(common.STATEINFORMEDDELTA)
			s.produceSignedHeartbeats()
			s.nextphase = "BlockGeneration"

		case "BlockGeneration":
			log.Print("STATE: BlockGeneration phase")
			s.produceBlock()
			s.nextphase = "BlockDecision"

		case "BlockDecision":
			log.Print("STATE: BlockDecision phase")
			ok := s.decideBlock()

			log.Print("STATE: ok=", ok)

			// If we accepted a block we should stop executing and let the
			// Next state take over
			if ok {
				s.nextphase = "Dead"
			} else {
				s.nextphase = "Heartbeat"
			}
		case "Dead":
			defer s.lock.Unlock()
			return

		default:
			panic("Unrecognized state: " + s.nextphase)
		}
		s.lock.Unlock()
	}
}

func (s *StateSwarmInformed) produceHeartbeat() {

	stage1, stage2 := common.HashedRandomData(sha256.New(), 8)
	s.stage2 = stage2

	h := NewHeartbeat(s.chain.Id, s.chain.Host, stage1, "")
	go s.sendUpdate(h)
}

func (s *StateSwarmInformed) produceSignedHeartbeats() {

	signatures := make(map[string]string)
	for id, _ := range s.heartbeats {
		//Dummy signature
		signatures[id] = id
	}

	h := NewHeartbeatList(s.chain.Id, s.chain.Host, s.heartbeats, signatures)
	go s.sendUpdate(h)
}

func (s *StateSwarmInformed) produceBlock() {

	b := NewBlock(s.chain.Id, s.chain.Host, s.heartbeats, s.signatures)
	go s.sendUpdate(b)
}

func (s *StateSwarmInformed) decideBlock() bool {
	log.Print("STATE: decideBlock")
	heartbeats := make(map[string]*Heartbeat)
	signatures := make(map[string]map[string]string)

	for id, heartbeat := range s.heartbeats {
		log.Print("Heartbeat id=", id, " len(signatures)=", len(s.signatures[id]))
		if len(s.signatures[id]) > common.SWARMSIZE/2 {
			heartbeats[id] = heartbeat
			signatures[id] = s.signatures[id]
		}
	}

	if len(heartbeats) <= common.SWARMSIZE/2 {
		return false
	}

	b := NewBlock(s.chain.Id, s.chain.Host, s.heartbeats, s.signatures)
	go func(b *Block) {
		s.chain.AddBlock(b)
		s.chain.SwitchState(NewStateSteady(s.chain, b, b.Heartbeats, s.stage2))
	}(b)

	return true

}

func (s *StateSwarmInformed) HandleUpdate(t common.Update) State {

	s.lock.Lock()
	defer s.lock.Unlock()

	switch n := t.(type) {
	case *Heartbeat:
		if s.nextphase != "HeartbeatSigning" &&
			s.nextphase != "Heartbeat" {
			return s
		}

		s.heartbeats[n.Id] = n
		log.Print("STATE: Node added")

	case *HeartbeatList:
		if s.nextphase != "BlockGeneration" &&
			s.nextphase != "HeartbeatSigning" {
			return s
		}

		for id, signature := range n.Signatures {
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
			return s
		}

		for id, signatures := range n.Signatures {
			if _, ok := s.heartbeats[id]; !ok {
				s.heartbeats[id] = n.Heartbeats[id]
			}
			if _, ok := s.signatures[id]; !ok {
				s.signatures[id] = make(map[string]string)
			}
			for host, signature := range signatures {
				s.signatures[id][host] = signature
			}
		}

	default:

	}
	return s
}
