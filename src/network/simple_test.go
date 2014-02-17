package network

import (
	"common"
	"testing"
)

func TestSimpleMultiplexer(t *testing.T) {

	c := make(chan common.NetworkObject)

	m := NewSimpleMultiplexer()
	m.AddListener("", c)

	go m.SendNetworkObject(common.NetworkObject{"", "", "", ""})
	_ = <-c

	c2 := make(chan common.NetworkObject)
	c3 := make(chan common.NetworkObject)

	m.AddListener("", c2)
	m.AddListener("", c3)

	go m.SendNetworkObject(common.NetworkObject{"", "", "", ""})

	_ = <-c
	_ = <-c2
	_ = <-c3

}
