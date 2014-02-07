package swarm

import (
	"testing"
)

func TestBlockMarshaling(t *testing.T) {
	b := new(Block)
	b.DRNGSeed = "1"
	b.Id = "2"
	b.Stage1Entropy = make(map[string]string)
	b.Stage1Entropy["Test"] = "1"
	b.Stage2Entropy = make(map[string]string)
	b.StorageMapping = make(map[string]interface{})

	s := b.MarshalString()

	b2, err := UnmarshalBlock(s)
	if err != nil {
		t.Fatal(err)
	}

	if b.DRNGSeed != b2.DRNGSeed {
		t.Fatal("DRNGSeed not equal")
	}

	if b.Id != b2.Id {
		t.Fatal("Id not equal")
	}

	if b.Stage1Entropy["Test"] != b2.Stage1Entropy["Test"] {
		t.Fatal("Stage1entropy not equal")
	}
}
