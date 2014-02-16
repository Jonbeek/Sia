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

}
