package swarm

import (
	"encoding/json"
	"log"
	"time"
)

type Block struct {
	Id         string
	Blockchain string
	Time       time.Time

	Heartbeats map[string]*Heartbeat
	Signatures map[string]map[string]string
}

func NewBlock(blockchain, id string, heartbeats map[string]*Heartbeat,
	signatures map[string]map[string]string, time time.Time) *Block {
	b := new(Block)
	b.Id = id
	b.Blockchain = blockchain
	b.Heartbeats = heartbeats
	b.Signatures = signatures
	b.Time = time
	return b
}

func (b *Block) SwarmId() string {
	return b.Blockchain
}

func (b *Block) UpdateId() string {
	return b.Id
}

func (b *Block) Type() string {
	return "Block"
}

func (b *Block) MarshalString() string {
	o, err := json.Marshal(b)
	if err != nil {
		log.Fatal("Unable to marshal Block:", err)
	}
	return string(o)
}
