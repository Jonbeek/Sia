package swarm

import (
	"common"
	"crypto/sha256"
	"time"
)

type State interface {
	HandleTransaction(t common.Transaction)
	HandleBlock(b *Block)
}

func newBlockChain(Host string, Id string, StorageMapping map[string]interface{}) (b *BlockChain) {
	b = new(BlockChain)
	b.Id = Id
	b.StorageMapping = StorageMapping
	b.outgoingTransactions = make(chan common.NetworkObject)
	return
}

func NewBlockChain(Host string, Id string, StorageMapping map[string]interface{}) (b *BlockChain) {
	b = newBlockChain(Host, Id, StorageMapping)
	b.state = NewStateSwarmInformed(b)
	return
}

func JoinBlockChain(Host string, Id string) (b *BlockChain) {
	b = newBlockChain(Host, Id, make(map[string]interface{}))
	//b.state = NewStateSwarmJoin(b)
	return
}

//List of states
// SwarmInformed - Swarm member shave been told to join swarm
// SwarmConnected - Swarm Members have succesfully formed a block
// SwarmLive - Swarm has sent a signal to the parent blockchain saying it is
//             alive and is in the steady state
// SwarmJoin - We are joining an already alive swarm
// SwarmDied - The swarm has died, terminate

type StateSwarmInformed struct {
	//Map of hosts seen to number of times they have failed to generate a block
	//Used for both host alive tracking & host block generation tracking
	hostsseen      map[string]int
	broadcastcount uint
	chain          *BlockChain
	blockgen       <-chan time.Time
}

func NewStateSwarmInformed(chain *BlockChain) (s *StateSwarmInformed) {
	s = new(StateSwarmInformed)
	s.chain = chain
	s.blockgen = time.Tick(5 * time.Second)
	go s.broadcastLife()
	go s.checkBlockGen()
	return
}

func (s *StateSwarmInformed) checkBlockGen() {
	for _ = range s.blockgen {

		//Dont't try to generate a block if we don't have a majority
		if len(s.hostsseen) < 128 {
			continue
		}

		hosts := make([]string, 0, len(s.hostsseen))

		//Pull all hosts who we haven't seen skipping a block
		for host, skipped := range s.hostsseen {
			if skipped != 0 {
				continue
			}
			hosts = append(hosts, host)
		}

		//Check if we should be the block generator
		compiler := common.RendezvousHash(sha256.New(), hosts, s.chain.Host)

		if compiler == s.chain.Host {
			//Compile block
		}
	}
}

func (s *StateSwarmInformed) broadcastLife() {
	s.broadcastcount += 1
	s.chain.outgoingTransactions <- common.TransactionNetworkObject(
		NewNodeAlive(s.chain.Host, s.chain.Id))
}

func (s *StateSwarmInformed) HandleTransaction(t common.Transaction) {
	switch n := t.(type) {
	case *NodeAlive:
		s.hostsseen[n.Node] = 0
		// Resend hostsseen count once you have seen a majority of hosts
		if len(s.hostsseen) > 128 && s.broadcastcount < 2 {
			s.broadcastLife()
		}
	default:
		return
	}
}

func (s *StateSwarmInformed) HandleBlock(b *Block) {

}
