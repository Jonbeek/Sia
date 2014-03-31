package network

import (
	"common"
	"common/log"
)

func NewSimpleMultiplexer() common.NetworkMultiplexer {
	s := new(SimpleMultiplexer)
	s.in = make(chan common.NetworkMessage)
	s.out = make(chan common.NetworkMessageHandler)
	go s.listen()
	return s
}

type SimpleMultiplexer struct {
	in    chan common.NetworkMessage
	out   chan common.NetworkMessageHandler
	Hosts []common.NetworkMessageHandler
}

func (s *SimpleMultiplexer) listen() {

	for {
		select {
		case c := <-s.out:
			log.Debugln("MULTI: Host added")
			s.Hosts = append(s.Hosts, c)
		case o := <-s.in:
			log.Debugln("MULTI: Transaction ", o, " to be sent to ", len(s.Hosts))
			for _, s := range s.Hosts {
				go s.HandleNetworkMessage(o)
			}
		}
	}
}

func (s *SimpleMultiplexer) AddListener(SwarmId string, c common.NetworkMessageHandler) {
	s.out <- c
}

func (s *SimpleMultiplexer) SendNetworkMessage(o common.NetworkMessage) {
	s.in <- o
}

func (s *SimpleMultiplexer) Listen(addr string) {
	panic("Not implemented")
}

func (s *SimpleMultiplexer) Connect(addr string) {
	panic("Not implemented")
}
