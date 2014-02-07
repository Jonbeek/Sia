package swarm

import (
	"encoding/json"
	"log"
)

var EntropyVolume int = 32

type Block struct {
	Id    string
	Swarm string

	EntropyStage1 map[string][]byte
	EntropyStage2 map[string][]byte

	//Mapping of hosts -> what they store
	StorageMapping map[string]interface{}
}

type State struct {
	DRNGSeed     []byte
	BlockHistory map[uint64]Block
}

func (b *Block) SwarmId() string {
	return b.Swarm
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
