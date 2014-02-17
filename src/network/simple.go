package network

import (
	"common"
	"log"
)

func NewSimpleMultiplexer() common.NetworkMultiplexer {
	in := make(chan common.NetworkObject)
	out := make(chan chan common.NetworkObject)
	s := &SimpleMultiplexer{in, out, nil}
	go s.listen()
	return s
}

type SimpleMultiplexer struct {
	in    chan common.NetworkObject
	out   chan chan common.NetworkObject
	Hosts []chan common.NetworkObject
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
				go func(s chan common.NetworkObject) {
					s <- o
					log.Println("MULTI: Transaction sent to host")
				}(s)
			}
			log.Println("MULTI: Finished Processing")
		}
		log.Println("MULTI: CYCLING")
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
