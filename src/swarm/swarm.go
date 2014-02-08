package swarm

import (
	"common"
	"time"
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

	SeenTransactions map[string]bool
}

func (b *BlockChain) AddSource(plexer common.NetworkMultiplexer) {

	c := make(chan common.NetworkObject)
	// plexer.AddListener(b.Id(), c)?
	go b.ReceiveObjects(c)
}

func (b *BlockChain) ReceiveObjects(c chan common.NetworkObject) {
	for o := range c {
		switch {
		case len(o.TransactionId) != 0:

			if b.SeenTransactions[o.TransactionId] {
				continue
			}

			b.SeenTransactions[o.TransactionId] = true

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

			//Verify BLock

			//Apply Block
			//Generate new heartbeat update
			// Figure out if I'm the block compiler?
			// if so, spawn a goroutine that will wait for 50% of the estimated
			// block time and run

			b.state.HandleBlock(block)

			return
		}

	}
}
