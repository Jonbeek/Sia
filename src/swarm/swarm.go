package swarm

type BlockChain struct {
	compiletime chan<- time.Time

	host map[string]bool
}

func (b *BlockChain) AddSource(plexer common.NetworkMultiplexer) {

	c := make(chan NetworkObject)
	go b.RecieveObjects(c)
}

func (b *BlockChain) RecieveObjects(c chan NetworkObject) {
	for o, _ := range c {
		switch {
		case len(o.TransactionId) != 0:
			//Decode Transaction
			//Process it if needed
			return

		case len(o.BlockId) != 0:
			//Verify BLock
			//Apply Block
			//Generate new heartbeat update
			// Figure out if I'm the block compiler?
			// if so, spawn a gorouting that will wait for 50% of the estimated
			// block time and run
			return
		}

	}
}
