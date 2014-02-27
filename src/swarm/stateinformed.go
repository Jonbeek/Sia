package swarm

import (
	"common"
	"crypto/sha256"
	"errors"
	"log"
	"sync"
	"time"
)

/* SwarmInformed

   This State initializes the arbitrary swarm. It uses a simple consensus
   algorithm.

   The first stage is the learning phase. Each node sends a NodeAlive message.
   They then wait to see at least a majority of the swarms NodeAlive messages
   then they resend their NodeAlive message, to make sure that the majority of
   the swarm knows they are alive.

   Once the learning period has expired, rendezvous hashing is used to select
   the host which should compile the block. The selected host creates a block
   and announces it. The hosts who agree that the selected host is the block
   compiler release a heartbeat update.

   Once the selected host has compiled heartbeats from the majority of swarms,
   the block compiler makes another block which indicates that the swarm is
   transitioning out of this state.

   If this process fails, then the host should be marked as an invalid compiler
   and the next block compiler attempted, until this process fails or some other
   timeout expires.

   TODO: This implementation handles failure after the first block has appeared
   badly. Also the learning phase -> block compiler transition should probably
   have a pause in it and we should have a constant start time, not just 1 second
   after we started a given structure.

   TODO: We can actually do stage1 hashes in the nodealive message and stage1
   and 2 in the heartbeat message. This would initialize us faster
*/
type StateSwarmInformed struct {
	//Map of hosts seen to number of times they have failed to generate a block
	//Used for both host alive tracking & host block generation tracking
	hostsseen map[string]int

	// How many times we have broadcast that we are alive, we use a two stage
	// process where we broadcast, and then broadcast again when we have seen
	// enough nodes up to form a majority
	broadcastcount uint
	stage2         string

	// This state has two phases, the learning phase where it listens for new
	// hosts and then the commit stage where it listens for a block that
	// is correct according to its knowledge and then votes for it.
	learning bool
	alive    bool

	heartbeats []*Heartbeat

	chain *Blockchain

	lock *sync.Mutex

	blockgen <-chan time.Time
	die      chan struct{}
}

type uwrap struct {
	update common.Update
	ret    chan State
}

func NewStateSwarmInformed(chain *Blockchain) (s *StateSwarmInformed) {
	s = new(StateSwarmInformed)
	s.chain = chain

	// When transitions / timeouts happen
	// Should be dynamically set
	s.blockgen = time.Tick(common.STATEINFORMEDDELTA)

	s.lock = &sync.Mutex{}

	s.learning = true
	s.alive = true
	s.hostsseen = make(map[string]int)
	s.die = make(chan struct{})

	go s.broadcastLife()
	go s.mainloop()
	return
}

func (s *StateSwarmInformed) Die(lock bool) {
	if lock {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	s.alive = false
}

func (s *StateSwarmInformed) sendUpdate(u common.Update) {
	s.chain.outgoingUpdates <- u
}

func (s *StateSwarmInformed) blockCompiler() (string, error) {

	hosts := make([]string, 0, len(s.hostsseen))

	//Pull all hosts who we haven't seen skipping a block
	for host, skipped := range s.hostsseen {
		if skipped != 0 {
			continue
		}
		hosts = append(hosts, host)
	}

	if len(hosts) == 0 {
		return "", errors.New("No valid hosts")
	}

	//Check if we should be the block generator
	return common.RendezvousHash(sha256.New(), hosts, s.chain.Host), nil
}

func (s *StateSwarmInformed) mainloop() {

	var compiler string
	var err error

	for {
		_ = <-s.blockgen
		s.lock.Lock()
		log.Print("STATE: Blockgen Recieved")

		if !s.alive {
			defer s.lock.Unlock()
			return
		}

		if s.learning {
			s.learning = false
			log.Print("STATE: Disable Learning")
		} else if len(s.chain.BlockHistory) == 0 {
			if compiler != "" {
				s.hostsseen[compiler] += 1
				log.Print("STATE: Block Compiler not found")
			}
		}

		//Dont't try to generate a block if we don't have a majority of hosts
		if len(s.hostsseen) <= common.SWARMSIZE/2 {
			s.lock.Unlock()
			continue
			//Should actually switch to state swarmdied after a while
		}

		compiler, err = s.blockCompiler()
		if err != nil {
			log.Print("State: ", err)
		}

		log.Print("STATE: Block Compiler ", compiler, " Me ", s.chain.Host,
			" chain len ", len(s.chain.BlockHistory),
			" good ", compiler == s.chain.Host)

		if len(s.chain.BlockHistory) == 0 && compiler == s.chain.Host {

			log.Print("STATE: Generating Block type 1")

			id, err := common.RandomString(8)
			if err != nil {
				panic("checkBlockGenRandom" + err.Error())
			}
			b := &Block{id, s.chain.Id, s.chain.Host, nil, nil}
			b.StorageMapping = make(map[string]interface{})
			for host, _ := range s.hostsseen {
				b.StorageMapping[host] = nil
			}

			s.heartbeats = s.heartbeats[0:0]
			time.Sleep(100 * time.Millisecond)
			go s.sendUpdate(b)
		}

		if len(s.chain.BlockHistory) == 1 && len(s.heartbeats) > 2 {

			log.Print("STATE: Generating Block type 2")

			id, err := common.RandomString(8)
			if err != nil {
				panic(err)
			}

			b := &Block{id, s.chain.Id, s.chain.Host, nil, nil}
			b.StorageMapping = s.chain.BlockHistory[0].StorageMapping
			b.Heartbeats = make(map[string]*Heartbeat)
			for _, h := range s.heartbeats {
				b.Heartbeats[h.Host] = h
			}

			// Arbitrary hard coded constant to make the testcases pass
			time.Sleep(500 * time.Millisecond)
			go s.sendUpdate(b)
		}
		s.lock.Unlock()
		log.Print("STATE: Signal Handling Finished")
	}
}

func (s *StateSwarmInformed) broadcastLife() {
	log.Print("STATE: NodeAlive Update to be Queed")
	s.lock.Lock()
	defer s.lock.Unlock()
	s.broadcastcount += 1
	go s.sendUpdate(NewNodeAlive(s.chain.Host, s.chain.Id))

}

func (s *StateSwarmInformed) HandleUpdate(t common.Update) State {
	s.lock.Lock()
	defer s.lock.Unlock()
	switch n := t.(type) {
	case *NodeAlive:
		if !s.learning {
			return s
		}

		s.hostsseen[n.Node] = 0
		// Resend hostsseen count once you have seen a majority of hosts
		if len(s.hostsseen) > 2 && s.broadcastcount < 2 {
			go s.broadcastLife()
		}

		log.Print("STATE: Node added")

	case *Heartbeat:

		//If we're learning this is too early
		if s.learning {
			return s
		}

		// If we're not trying to compile we don't care
		c, _ := s.blockCompiler()
		if c != s.chain.Host {
			return s
		}

		if len(s.chain.BlockHistory) != 0 && n.ParentBlock == s.chain.BlockHistory[0].Id {
			s.heartbeats = append(s.heartbeats, n)
		}

	case *Block:
		return s.handleBlock(n)

	default:
	}
	return s
}

func (s *StateSwarmInformed) handleBlock(b *Block) State {

	// If the learning timeout hasn't expired, don't accept blocks
	if s.learning {
		log.Print("STATE: Block rejected b/c learning")
		return s
	}

	// All blocks in this state should be generated by the ideal host
	c, _ := s.blockCompiler()
	if b.Compiler != c {
		log.Print("STATE: Block rejected b/c wrong compiler")
		return s
	}

	// We are looking for a block to generate a heartbeat for
	if len(s.chain.BlockHistory) == 0 {
		log.Print("STATE: Block accepted type 1")
		s.chain.AddBlock(b)

		stage1, stage2 := common.HashedRandomData(sha256.New(), 8)
		s.stage2 = stage2

		if _, ok := b.StorageMapping[s.chain.Host]; ok {
			h := NewHeartbeat(s.chain.BlockHistory[0], s.chain.Host, stage1, "")
			go s.sendUpdate(h)
		}
		return s
	}

	// We're looking for the block with heartbeats to figure out if we're in
	// it
	if len(s.chain.BlockHistory) == 1 {
		log.Print("STATE: Block accepted type 2")

		s.chain.AddBlock(b)

		if _, ok := b.StorageMapping[s.chain.Host]; ok {
			//If we're in the block switch to signal mode
			log.Print("STATE: Switching to connected")
			s.Die(false)
			return NewStateSteady()
		} else {
			//Join the swarm
			log.Print("STATE: Switching to Join")
		}
	}

	return s
}
