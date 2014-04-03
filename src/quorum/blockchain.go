package quorum

import (
	"common"
	"common/log"
	"sync"
)

type Blockchain struct {
	Host  string
	Id    string
	state State

	incomingMessages chan common.Message
	outgoingUpdates  chan common.Update

	blockHistory []*Block

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
		log.Debug("SWARM: sending outgoing transaction")
		plexer.SendMessage(common.MarshalUpdate(i))
	}
}

func (b *Blockchain) HandleMessage(m common.Message) {
	log.Debug("SWARM: network message recieved")

	u, err := UnmarshalUpdate(m)
	if err != nil {
		panic(err)
	}

	b.state.HandleUpdate(u)
	log.Debug("SWARM: Update handling finished")
}

func (b *Blockchain) AddBlock(block *Block) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.blockHistory != nil && len(b.blockHistory) == 5 {
		b.blockHistory = b.blockHistory[:4]
	}
	b.blockHistory = append(b.blockHistory, block)
}

func (b *Blockchain) LastBlock() *Block {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.blockHistory[len(b.blockHistory)-1]
}

func (b *Blockchain) BlockLen() int {
	b.lock.Lock()
	defer b.lock.Unlock()
	return len(b.blockHistory)
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

func (b *Blockchain) HostActive(host string) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	_, ok := b.blockHistory[len(b.blockHistory)-1].Heartbeats[host]
	return ok
}
