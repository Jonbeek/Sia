package swarm

import (
	"common"
	"time"
)

type BlockChain struct {
	Id          string
	state       uint
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
	// plexer.AddListener(b.Id(), c)?
	go b.ReceiveObjects(c)
}

func (b *BlockChain) ReceiveObjects(c chan common.NetworkObject) {
	for o := range c {
		switch {
		case len(o.TransactionId) != 0:
			//Decode Transaction
			//Process it if needed
			// For now, store all transactions
			// append(b.transactions, o)
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
			_ = block

			//Apply Block
			//Generate new heartbeat update
			// Figure out if I'm the block compiler?
			// if so, spawn a goroutine that will wait for 50% of the estimated
			// block time and run
			return
		}

	}
}
