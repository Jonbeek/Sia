package swarm

import (
	"testing"
)

func TestBlockMarshaling(t *testing.T) {
	b := new(Block)
	b.Id = "2"
	b.Heartbeats = make(map[string]Heartbeat)
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
