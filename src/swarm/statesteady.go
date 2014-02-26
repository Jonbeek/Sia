package swarm

import (
	"common"
	"crypto/sha256"
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
	Heartbeats   []*Heartbeat
	curHeartbeat *Heartbeat

	blocklock sync.Mutex
}

func NewStateSteady(chain *Blockchain, block *Block, hostsseen map[string]int) (s *StateSteady) {
	s = new(StateSteady)
	s.chain = chain
	s.block = block
	s.blocksend = make(chan string)
	s.update = make(chan common.Update, 1)
	s.Hosts = make(map[string]bool)

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
			s.blocklock.Lock()
			if b == s.block.UpdateId() {
				// Should take into account delinquent block compilers
				s.compileBlock()
			}
			s.blocklock.Unlock()
		case <-s.die:
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
	default:
		// Do nothing, for now.
	}
}

func (s *StateSteady) handleBlock(b *Block) {
	s.blocklock.Lock()
	defer s.blocklock.Unlock()
	go s.makeHeartbeat(s.block)
	// VerifyBlock(b)
	s.block = b
	s.chain.AddBlock(b)
	go func(Id string) {
		time.Sleep(5 * time.Second)
		s.blocksend <- Id
	}(s.block.UpdateId())
	s.Heartbeats = s.Heartbeats[0:0]
}

func (s *StateSteady) handleHeartbeat(h *Heartbeat) {
	// Record Stage2
	// Increment count of heartbeats
	// If count equal to number of connected hosts
	// Initiate block compilation
	// New round of heartbeats.
	s.Heartbeats = append(s.Heartbeats, h)

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
	compiler := common.RendezvousHash(sha256.New(), hosts, s.block.UpdateId())
	// If we are the block compiler, make a block and send and handle it
	// Otherwise, do nothing and wait for the block compiler to send a block
	if compiler == s.chain.Host {
		id, _ := common.RandomString(8)
		b := &Block{id, s.chain.Id, s.chain.Host, nil, nil}
		b.StorageMapping = s.chain.BlockHistory[0].StorageMapping
		b.Heartbeats = make(map[string]*Heartbeat)
		for _, h := range s.Heartbeats {
			b.Heartbeats[h.Host] = h
		}
		go s.sendUpdate(b)
	}
}

func (s *StateSteady) makeHeartbeat(prevState *Block) {
	var Stage1, Stage2 string
	Stage2 = s.secretstring
	Stage1, s.secretstring = common.HashedRandomData(sha256.New(), 8)
	s.curHeartbeat = NewHeartbeat(prevState, Stage1, Stage2)
	s.sendUpdate(s.curHeartbeat)
}

func (s *StateSteady) sendUpdate(u common.Update) {
	s.chain.outgoingUpdates <- u
}
