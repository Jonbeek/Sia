package network

import (
	"common"
	"testing"
)

func TestNetworkMultiplexer(t *testing.T) {

	c := newTestListener()

	m := NewNetworkMultiplexer()
	m.AddListener("", c)

	go m.SendNetworkMessage(common.NetworkMessage{"", "", "", ""})
	c.Receive()

	c2 := newTestListener()
	c3 := newTestListener()
	c4 := newTestListener()

	m.AddListener("1", c2)
	m.AddListener("2", c3)
	m.AddListener("2", c4)

	go m.SendNetworkMessage(common.NetworkMessage{"1", "", "", ""})

	c2.Receive()

	go m.SendNetworkMessage(common.NetworkMessage{"2", "", "", ""})

	c3.Receive()
	c4.Receive()

	n := NewNetworkMultiplexer()

	go m.Listen(":1234")
	go n.Connect(":1234")

	o := NewNetworkMultiplexer()

	go o.Connect(":1234")
}
