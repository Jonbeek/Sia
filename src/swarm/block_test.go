package swarm

import (
	"testing"
)

func TestBlockMarshaling(t *testing.T) {
	b := new(Block)
	b.Id = "2"
	b.EntropyStage1 = make(map[string][]byte)
	b.EntropyStage2 = make(map[string][]byte)
	b.StorageMapping = make(map[string]interface{})

	s := b.MarshalString()

	b2, err := UnmarshalBlock(s)
	if err != nil {
		t.Fatal(err)
	}

	if b.Id != b2.Id {
		t.Fatal("Id not equal")
	}
}
