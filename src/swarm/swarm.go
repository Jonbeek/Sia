package swarm

import (
	"common"
	"time"
)

type BlockChain struct {
	compiletime chan<- time.Time

	host map[string]bool
	// transactions []common.Transaction
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
			//Verify BLock
			//Apply Block
			//Generate new heartbeat update
			// Figure out if I'm the block compiler?
			// if so, spawn a goroutine that will wait for 50% of the estimated
			// block time and run
			return
		}

	}
}
