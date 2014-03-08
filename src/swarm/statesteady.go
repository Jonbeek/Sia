package swarm

import (
	"common"
	"crypto/sha256"
	"log"
	"sync"
	"time"
)

// StateSteady is the standard state of the swarm

type StateSteady struct {
	// blocksend is a timeout channel, with some modifications so
	// previous heartbeat timeout doesn't affect current heartbeat
	blocksend chan string
	update    chan common.Update
	die       chan struct{}

	// Hosts is a record of all hosts in the swarm
	Hosts map[string]bool

	chain *Blockchain
	// Current block, all Hosts will be creating new block.
	block        *Block
	secretstring string
	// Heartbeats is all received heartbeats from other hosts.
	Heartbeats map[string]*Heartbeat

	lock sync.Mutex
}

func NewStateSteady(chain *Blockchain, block *Block, hostsseen map[string]int, secretstring string) (s *StateSteady) {
	log.Println("STATE: Creating new StateSteady")
	s = new(StateSteady)
	s.chain = chain
	s.block = block
	s.blocksend = make(chan string)
	s.update = make(chan common.Update, 1)
	s.Hosts = make(map[string]bool)
	s.secretstring = secretstring
	s.Heartbeats = make(map[string]*Heartbeat)

	for k, _ := range hostsseen {
		s.Hosts[k] = true
	}

	go s.mainloop()
	go s.makeHeartbeat(block)
	return
}

func (s *StateSteady) mainloop() {
	for {
		select {
		case u := <-s.update:
			s.handleUpdate(u)
		case b := <-s.blocksend:
			// Host did not receive all heartbeats in time
			// Use current set of heartbeats to create new Block
			s.lock.Lock()
			if b == s.block.UpdateId() {
				log.Print("STATE: Timed out waiting for heartbeats on block ID ", b)
				// Should take into account delinquent block compilers
				s.compileBlock()
			}
			s.lock.Unlock()
		case <-s.die:
			// Kill Channel
			return
		}
	}
}

func (s *StateSteady) HandleUpdate(u common.Update) State {
	s.update <- u
	return s
}

func (s *StateSteady) handleUpdate(u common.Update) {
	switch n := u.(type) {
	case *Heartbeat:
		s.handleHeartbeat(n)
	case *Block:
		s.handleBlock(n)
	case *NodeAlive:
		log.Println("STATE: Steady received NodeAlive: Blockchain ", n.SwarmId(), ", Id ", n.UpdateId())
	default:
		// Only recording type and Blockchain source.
		log.Println("STATE: Steady received unknown Update: Type ", n.Type(), ", Blockchain ", n.SwarmId())
	}
}

func (s *StateSteady) handleBlock(b *Block) {
	// Verify block, update block used, send new heartbeats
	log.Println("STATE: Steady handling new Block")
	s.lock.Lock()
	defer s.lock.Unlock()
	// VerifyBlock(b)
	go s.makeHeartbeat(s.block)
	s.block = b
	s.chain.AddBlock(b)
	go func(Id string) {
		time.Sleep(5 * time.Second)
		s.blocksend <- Id
	}(s.block.UpdateId())
	s.Heartbeats = make(map[string]*Heartbeat)
}

func (s *StateSteady) handleHeartbeat(h *Heartbeat) {
	// See if heartbeat from unseen host
	// Record heartbeat.
	// If heartbeat count equal to count of hosts, initiate block compilation
	log.Println("STATE: Steady handling Heartbeat: Blockchain ", h.SwarmId(), ", Id ", h.UpdateId())
	if _, ok := s.Heartbeats[h.Host]; !ok {
		s.Heartbeats[h.Host] = h
	}
	if len(s.Heartbeats) == len(s.Hosts) {
		s.compileBlock()
	}
}

func (s *StateSteady) compileBlock() {
	var hosts []string
	// the set of hosts is stored as a map for easy insertion + deletion
	for host, _ := range s.Hosts {
		hosts = append(hosts, host)
	}
	compiler := common.RendezvousHash(sha256.New(), hosts, string(s.chain.DRNGSeed))
	log.Println("STATE: Steady block compiler is ", compiler)
	// If we are the block compiler, make a block and send and handle it
	// Otherwise, do nothing and wait for the block compiler to send a block
	if compiler == s.chain.Host {
		log.Println("STATE: ", compiler, " creating new block")
		id, _ := common.RandomString(8)
		b := &Block{id, s.chain.Id, s.chain.Host, nil, nil}
		b.StorageMapping = s.block.StorageMapping
		b.Heartbeats = make(map[string]*Heartbeat)
		for s, h := range s.Heartbeats {
			b.Heartbeats[s] = h
		}
		go s.sendUpdate(b)
	}
}

func (s *StateSteady) makeHeartbeat(prevState *Block) {
	// Create new heartbeat using previous, then send to Blockchain
	log.Println("STATE: Making new round of heartbeats")
	var Stage1, Stage2 string
	Stage2 = s.secretstring
	Stage1, s.secretstring = common.HashedRandomData(sha256.New(), 8)
	curHeartbeat := NewHeartbeat(prevState, s.chain.Host, Stage1, Stage2)
	s.sendUpdate(curHeartbeat)
}

func (s *StateSteady) sendUpdate(u common.Update) {
	// Blockchain handles this, pass update to the blockchain
	s.chain.outgoingUpdates <- u
}
