package network

import (
	"common"
	"log"
)

func NewSimpleMultiplexer() common.NetworkMultiplexer {
	in := make(chan common.NetworkMessage)
	out := make(chan chan common.NetworkMessage)
	s := &SimpleMultiplexer{in, out, nil}
	go s.listen()
	return s
}

type SimpleMultiplexer struct {
	in    chan common.NetworkMessage
	out   chan chan common.NetworkMessage
	Hosts []chan common.NetworkMessage
}

func (s *SimpleMultiplexer) listen() {

	for {
		select {
		case c := <-s.out:
			log.Println("MULTI: Host added")
			s.Hosts = append(s.Hosts, c)
		case o := <-s.in:
			log.Println("MULTI: Transaction ", o, " to be sent to ", len(s.Hosts))
			for _, s := range s.Hosts {
				go func(s chan common.NetworkMessage) {
					s <- o
					log.Println("MULTI: Transaction sent to host")
				}(s)
			}
			log.Println("MULTI: Finished Processing")
		}
		log.Println("MULTI: CYCLING")
	}
}

func (s *SimpleMultiplexer) AddListener(SwarmId string, c chan common.NetworkMessage) {
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
