package network

import (
	"common"
)

func NewSimpleMultiplexer() common.NetworkMultiplexer {
	in := make(chan common.NetworkObject)
	out := make(chan chan common.NetworkObject)
	s := &SimpleMultiplexer{in, out}
	go s.listen(in, out)
	return s
}

type SimpleMultiplexer struct {
	in  chan common.NetworkObject
	out chan chan common.NetworkObject
}

func (s *SimpleMultiplexer) listen(in chan common.NetworkObject,
	out chan chan common.NetworkObject) {

	hosts := make([]chan common.NetworkObject, 0)

	for {
		select {
		case c := <-out:
			hosts = append(hosts, c)
		case o := <-in:
			for _, s := range hosts {
				s <- o
			}
		}
	}
}

func (s *SimpleMultiplexer) AddListener(SwarmId string, c chan common.NetworkObject) {
	s.out <- c
}

func (s *SimpleMultiplexer) SendNetworkObject(o common.NetworkObject) {
	s.in <- o
}

func (s *SimpleMultiplexer) Listen(addr string) {
	panic("Not implemented")
}

func (s *SimpleMultiplexer) Connect(addr string) {
	panic("Not implemented")
}
