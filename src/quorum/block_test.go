package quorum

import (
	"common"
	"testing"
)

func TestBlockMarshaling(t *testing.T) {
	b := new(Block)
	b.Id = "2"
	b.Heartbeats = make(map[string]*Heartbeat)

	s := common.MarshalUpdate(b)

	b2, err := UnmarshalUpdate(s)
	if err != nil {
		t.Fatal(err)
	}

	if b.Id != b2.(*Block).Id {
		t.Fatal("Id not equal")
	}
}
