package network

import (
	"common"
	"testing"
)

func TestSimpleMultiplexer(t *testing.T) {

	c := make(chan common.NetworkMessage)

	m := NewSimpleMultiplexer()
	m.AddListener("", c)

	go m.SendNetworkMessage(common.NetworkMessage{"", "", "", ""})
	_ = <-c

	c2 := make(chan common.NetworkMessage)
	c3 := make(chan common.NetworkMessage)

	m.AddListener("", c2)
	m.AddListener("", c3)

	go m.SendNetworkMessage(common.NetworkMessage{"", "", "", ""})

	_ = <-c
	_ = <-c2
	_ = <-c3

}
