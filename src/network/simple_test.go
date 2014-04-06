package network

import (
	"common"
	"testing"
)

type testListener chan common.Message

func newTestListener() (l testListener) {
	l = make(chan common.Message)
	return
}

func (l testListener) HandleMessage(m common.Message) {
	l <- m
}

func (l testListener) Receive() {
	_ = <-l
}

func TestSimpleMultiplexer(t *testing.T) {

	c := newTestListener()

	m := NewSimpleMultiplexer()
	m.AddListener("", c)

	go m.SendMessage(common.Message{"1", ""})
	c.Receive()

	c2 := newTestListener()
	c3 := newTestListener()

	m.AddListener("", c2)
	m.AddListener("", c3)

	go m.SendMessage(common.Message{"2", ""})

	c.Receive()
	c2.Receive()
	c3.Receive()

}
