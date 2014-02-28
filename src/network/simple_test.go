package network

import (
	"common"
	"testing"
)

type testListener chan common.NetworkMessage

func newTestListener() (l testListener) {
	l = make(chan common.NetworkMessage)
	return
}

func (l testListener) HandleNetworkMessage(m common.NetworkMessage) {
	l <- m
}

func (l testListener) Receive() {
	_ = <-l
}

func TestSimpleMultiplexer(t *testing.T) {

	c := newTestListener()

	m := NewSimpleMultiplexer()
	m.AddListener("", c)

	go m.SendNetworkMessage(common.NetworkMessage{"", "", "", ""})
	c.Receive()

	c2 := newTestListener()
	c3 := newTestListener()

	m.AddListener("", c2)
	m.AddListener("", c3)

	go m.SendNetworkMessage(common.NetworkMessage{"", "", "", ""})

	c.Receive()
	c2.Receive()
	c3.Receive()

}
