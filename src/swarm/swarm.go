package swarm

import (
	"common"
	"time"
)

const (
	EntropyVolume = 32
)

type BlockChain struct {
	Host        string
	Id          string
	state       State
	compiletime chan<- time.Time

	outgoingTransactions chan common.NetworkObject

	// transactions []common.Transaction
	BlockHistory []*Block

	//Updated every block
	DRNGSeed       []byte
	StorageMapping map[string]interface{}
}

func (b *BlockChain) AddSource(plexer common.NetworkMultiplexer) {

	c := make(chan common.NetworkObject)
	go plexer.AddListener(b.Id, c)
	go b.ReceiveObjects(c)
	go b.SendObjects(plexer)
}

func (b *BlockChain) SendObjects(plexer common.NetworkMultiplexer) {
	for i := range b.outgoingTransactions {
		plexer.SendNetworkObject(i)
	}
}

func (b *BlockChain) ReceiveObjects(c chan common.NetworkObject) {
	SeenTransactions := make(map[string]bool)
	for o := range c {
		switch {
		case len(o.TransactionId) != 0:

			if SeenTransactions[o.TransactionId] {
				continue
			}

			SeenTransactions[o.TransactionId] = true

			t, err := UnmarshalTransaction(o.Payload)
			if err != nil {
				continue
			}

			b.state.HandleTransaction(t)

			return

		case len(o.BlockId) != 0:

			if o.BlockId == b.BlockHistory[0].Id {
				continue
			}

			block, err := UnmarshalBlock(o.Payload)
			if err != nil {
				continue
			}

			b.state = b.state.HandleBlock(block)

			return
		}
	}
}

func (b *BlockChain) AddBlock(block *Block) {
	if b.BlockHistory != nil {
		b.BlockHistory = b.BlockHistory[:4]
	}
	b.BlockHistory = append(b.BlockHistory, block)
}
