package swarm

import (
	"common"
	"log"
	"sync"
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

	lock *sync.Mutex
}

func (b *Blockchain) AddSource(plexer common.NetworkMultiplexer) {

	plexer.AddListener(b.Id, b)
	go b.mainloop(plexer)
}

func (b *Blockchain) mainloop(plexer common.NetworkMultiplexer) {
	for i := range b.outgoingUpdates {
		log.Print("SWARM: sending outgoing transaction")
		plexer.SendNetworkMessage(common.MarshalUpdate(i))
	}
}

func (b *Blockchain) HandleNetworkMessage(m common.NetworkMessage) {
	b.lock.Lock()
	defer b.lock.Unlock()

	log.Print("SWARM: network message recieved")

	if b.SeenTransactions[m.UpdateId] {
		log.Print("Swarm: Update Already Seen")
		return
	}

	u, err := UnmarshalUpdate(m)
	if err != nil {
		panic(err)
	}

	b.state = b.state.HandleUpdate(u)
	log.Print("SWARM: Update handling finished")
}

func (b *Blockchain) AddBlock(block *Block) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.BlockHistory != nil && len(b.BlockHistory) == 5 {
		b.BlockHistory = b.BlockHistory[:4]
	}
	b.BlockHistory = append(b.BlockHistory, block)
}

func (b *Blockchain) LastBlock() *Block {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.BlockHistory[len(b.BlockHistory)-1]
}

func (b *Blockchain) BlockLen() int {
	b.lock.Lock()
	defer b.lock.Unlock()
	return len(b.BlockHistory)
}

func (b *Blockchain) SwitchState(s State) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.state = s
}
func (b *Blockchain) GetState() State {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.state
}
