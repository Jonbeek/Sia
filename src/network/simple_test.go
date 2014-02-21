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

func (l testListener) Recieve() {
	_ = <-l
}

func TestSimpleMultiplexer(t *testing.T) {

	c := newTestListener()

	m := NewSimpleMultiplexer()
	m.AddListener("", c)

	go m.SendNetworkMessage(common.NetworkMessage{"", "", "", ""})
	c.Recieve()

	c2 := newTestListener()
	c3 := newTestListener()

	m.AddListener("", c2)
	m.AddListener("", c3)

	go m.SendNetworkMessage(common.NetworkMessage{"", "", "", ""})

	c.Recieve()
	c2.Recieve()
	c3.Recieve()

}
