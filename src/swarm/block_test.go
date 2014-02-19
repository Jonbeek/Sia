package swarm

import (
	"testing"
)

func TestIntegrate(t *testing.T) {

}

func Testverify(t *testing.T) {

}

// Incomplete
func TestMarshaling(t *testing.T) {
	b := new(Block)
	b.Id = "2"
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
