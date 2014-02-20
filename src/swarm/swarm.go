package swarm

import (
	"common"
	"log"
	"time"
)

type Blockchain struct {
	Host        string
	Id          string
	state       State
	compiletime chan<- time.Time

	outgoingTransactions chan common.NetworkMessage

	// transactions []common.Transaction
	BlockHistory []*Block

	//Updated every block
	DRNGSeed         []byte
	StorageMapping   map[string]interface{}
	SeenTransactions map[string]bool
}

func (b *Blockchain) AddSource(plexer common.NetworkMultiplexer) {

	c := make(chan common.NetworkMessage)
	plexer.AddListener(b.Id, c)
	go b.mainloop(plexer, c)
}

func (b *Blockchain) mainloop(plexer common.NetworkMultiplexer, c chan common.NetworkMessage) {
	log.Print("SWARM: mainloop started")
	for {
		log.Print("SWARM: Mainloop waiting for event", b.Host)
		select {
		case i := <-b.outgoingTransactions:
			log.Print("SWARM: sending outgoing transaction")
			plexer.SendNetworkMessage(i)
			log.Print("SWARM: Object sent")
		case o := <-c:
			log.Print("SWARM: network object recieved")
			switch {
			case len(o.TransactionId) != 0:

				log.Print("SWARM: Tis Transaction")
				if b.SeenTransactions[o.TransactionId] {
					log.Print("Swarm: Transaction Already Seen")
					continue
				}

				b.SeenTransactions[o.TransactionId] = true

				t, err := UnmarshalTransaction(o.Payload)
				if err != nil {
					panic(err)
				}

				b.state.HandleTransaction(t)
				log.Print("SWARM: Transaction handling finished")

			case len(o.BlockId) != 0:
				log.Print("SWARM: Block Recieved")

				block, err := UnmarshalBlock(o.Payload)
				if err != nil {
					continue
				}

				b.state = b.state.HandleBlock(block)

			}
		}
	}
}

func (b *Blockchain) AddBlock(block *Block) {
	if b.BlockHistory != nil && len(b.BlockHistory) == 5 {
		b.BlockHistory = b.BlockHistory[:4]
	}
	b.BlockHistory = append(b.BlockHistory, block)
}
