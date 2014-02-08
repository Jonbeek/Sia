package swarm

import (
	"common"
)

type State interface {
	HandleTransaction(t common.Transaction)
	HandleBlock(b *Block)
}

type StateSwarmInformed struct {
	hostsseen      map[string]bool
	broadcastcount uint
	chain          *BlockChain
}

func NewStateSwarmInformed(chain *BlockChain) (s *StateSwarmInformed) {
	s = new(StateSwarmInformed)
	s.chain = chain
	go s.broadcastLife()
	return
}

func (s *StateSwarmInformed) broadcastLife() {
	s.broadcastcount += 1
	s.chain.outgoingTransactions <- common.TransactionNetworkObject(
		NewNodeAlive(s.chain.Host, s.chain.Id))
}

func (s *StateSwarmInformed) HandleTransaction(t common.Transaction) {
	switch n := t.(type) {
	case *NodeAlive:
		s.hostsseen[n.Node] = true
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

const (
	// A majority of the swarm members have sent heartbeats, form initial block
	SWARM_CONNECTED = iota
	// The parent layer has been told the swarm is initialised, SteadyState
	SWARM_LIVE
	// The swarm already exists and we are joining it
	SWARM_JOIN
	// The swarm died
	SWARM_DIED
)

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
