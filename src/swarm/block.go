package swarm

import (
	"encoding/json"
	"log"
)

type Block struct {
	Id         string
	Blockchain string
	Compiler   string

	Heartbeats map[string]*Heartbeat

	//Mapping of hosts -> what they store
	StorageMapping map[string]interface{}
}

func (b *Block) SwarmId() string {
	return b.Blockchain
}

func (b *Block) BlockId() string {
	return b.Id
}

func (b *Block) MarshalString() string {
	o, err := json.Marshal(b)
	if err != nil {
		log.Fatal("Unable to marshal Block:", err)
	}
	return string(o)
}

func UnmarshalBlock(encoded string) (b *Block, err error) {
	b = new(Block)
	err = json.Unmarshal([]byte(encoded), b)
	return b, err
}
