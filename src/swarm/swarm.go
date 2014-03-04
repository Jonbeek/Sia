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

	incomingMessages chan common.NetworkMessage
	outgoingUpdates  chan common.Update

	// transactions []common.Transaction
	BlockHistory []*Block

	//Updated every block
	DRNGSeed         []byte
	StorageMapping   map[string]interface{}
	SeenTransactions map[string]bool

	//mapping of the wallets
	WalletMapping map[string]uint64
}

func (b *Blockchain) AddSource(plexer common.NetworkMultiplexer) {

	plexer.AddListener(b.Id, b)
	go b.mainloop(plexer)
}

func (b *Blockchain) mainloop(plexer common.NetworkMultiplexer) {
	for {
		select {
		case i := <-b.outgoingUpdates:
			log.Print("SWARM: sending outgoing transaction")
			plexer.SendNetworkMessage(common.MarshalUpdate(i))
		case m := <-b.incomingMessages:
			log.Print("SWARM: network message recieved")

			if b.SeenTransactions[m.UpdateId] {
				log.Print("Swarm: Update Already Seen")
				continue
			}

			u, err := UnmarshalUpdate(m)
			if err != nil {
				panic(err)
			}

			b.state = b.state.HandleUpdate(u)
			log.Print("SWARM: Update handling finished")
		}
	}
}

func (b *Blockchain) AddBlock(block *Block) {
	if b.BlockHistory != nil && len(b.BlockHistory) == 5 {
		b.BlockHistory = b.BlockHistory[:4]
	}
	b.BlockHistory = append(b.BlockHistory, block)
}

func (b *Blockchain) HandleNetworkMessage(m common.NetworkMessage) {
	b.incomingMessages <- m
}
