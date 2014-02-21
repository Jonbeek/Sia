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

	outgoingMessages chan common.NetworkMessage
	incomingMessages chan common.NetworkMessage

	// transactions []common.Transaction
	BlockHistory []*Block

	//Updated every block
	DRNGSeed         []byte
	StorageMapping   map[string]interface{}
	SeenTransactions map[string]bool
}

func (b *Blockchain) AddSource(plexer common.NetworkMultiplexer) {

	plexer.AddListener(b.Id, b)
	go b.mainloop(plexer)
}

func (b *Blockchain) mainloop(plexer common.NetworkMultiplexer) {
	for {
		select {
		case i := <-b.outgoingMessages:
			log.Print("SWARM: sending outgoing networkmessage")
			plexer.SendNetworkMessage(i)
		case m := <-b.incomingMessages:
			switch {
			case len(m.TransactionId) != 0:

				log.Print("SWARM: Transaction Recieved")
				if b.SeenTransactions[m.TransactionId] {
					log.Print("Swarm: Transaction Already Seen")
					continue
				}

				b.SeenTransactions[m.TransactionId] = true

				t, err := UnmarshalTransaction(m.Payload)
				if err != nil {
					panic(err)
				}

				b.state.HandleTransaction(t)

			case len(m.BlockId) != 0:
				log.Print("SWARM: Block Recieved")

				block, err := UnmarshalBlock(m.Payload)
				if err != nil {
					continue
				}

				b.state = b.state.HandleBlock(block)

			default:
				panic("Empty network message??")
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

func (b *Blockchain) HandleNetworkMessage(m common.NetworkMessage) {
	b.incomingMessages <- m
}
