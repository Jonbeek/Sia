package swarm

import (
	"encoding/json"
	"log"
)

type Block struct {
	DRNGSeed string
	Id       string
	Swarm    string

	Stage1Entropy map[string]string
	// Note DRNGSeed = sha512(Sorted(Array(Stage2Entropy)))
	Stage2Entropy map[string]string

	//Mapping of hosts -> what they store
	StorageMapping map[string]interface{}
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
