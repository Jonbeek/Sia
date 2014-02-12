package swarm

import (
	"common"
)

const (
	// Starting state, swarm members should start listening
	SWARM_INFORMED = iota
	// A majority of the swarm members have sent heartbeats
	SWARM_CONNECTED
	// The parent layer has been told the swarm is initialised, SteadyState
	SWARM_LIVE
	// The swarm already exists and we are joining it
	SWARM_JOIN
	// The swarm died
	SWARM_DIED
)

func newBlockChain(Id string, StorageMapping map[string]interface{}) (b *BlockChain) {
	b = new(BlockChain)
	b.Id = Id
	b.StorageMapping = StorageMapping
	b.outgoingTransactions = make(chan common.NetworkObject)
	return
}

func NewBlockChain(Host string, Id string, StorageMapping map[string]interface{}) (b *BlockChain) {
	b = newBlockChain(Id, StorageMapping)
	b.state = SWARM_INFORMED
	go func() {
		b.outgoingTransactions <- common.TransactionNetworkObject(NewNodeAlive(Host, Id))
	}()
	return
}

func JoinBlockChain(Host string, Id string) (b *BlockChain) {
	b = newBlockChain(Id, make(map[string]interface{}))
	b.state = SWARM_JOIN
	go func() {
		b.outgoingTransactions <- common.TransactionNetworkObject(NewNodeAlive(Host, Id))
	}()
	return
}
