package swarm

import (
	"common"
)

//List of states
// SwarmInformed - Swarm member shave been told to join swarm
// SwarmConnected - Swarm Members have succesfully formed a block
// SwarmLive - Swarm has sent a signal to the parent blockchain saying it is
//             alive and is in the steady state
// SwarmJoin - We are joining an already alive swarm
// SwarmDied - The swarm has died, terminate

type State interface {
	HandleTransaction(t common.Transaction)
	HandleBlock(b *Block) State
}

type StateSteady struct {
	// swarm information - id, network location, etc.
	// connection information to all hosts in swarm
	// connection information to all hosts in parent swarm

	// wallet database

	// block history, snapshot history (if there is a snapshot history)

	// scheduled scripts (once we write those in)

	// active heartbeat - being updated by transactions and such

	DRNGSeed string

	// The data used to produce Stage1 hashes in the recent heartbeat
	SecretEntropy   string
	SecretFileProof string

	ActiveBlock Block
}

func newBlockChain(Host string, Id string, StorageMapping map[string]interface{}) (b *BlockChain) {
	b = new(BlockChain)
	b.Host = Host
	b.Id = Id
	b.StorageMapping = StorageMapping
	b.outgoingTransactions = make(chan common.NetworkObject)
	b.SeenTransactions = make(map[string]bool)
	return
}

func NewBlockChain(Host string, Id string, StorageMapping map[string]interface{}) (b *BlockChain) {
	b = newBlockChain(Host, Id, StorageMapping)
	b.state = NewStateSwarmInformed(b)
	return
}

func JoinBlockChain(Host string, Id string) (b *BlockChain) {
	b = newBlockChain(Host, Id, make(map[string]interface{}))
	b.state = NewStateSwarmJoin(b)
	return
}
