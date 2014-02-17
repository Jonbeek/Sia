package network

import (
	"common"
	//	"log"
)

func NewNetworkMultiplexer() *NetworkMultiplexer {
	in := make(map[string][]chan common.NetworkObject)
	out := make(map[string][]chan chan common.NetworkObject)
	m := &NetworkMultiplexer{in, out, nil}
	go m.listen()
	return m
}

//Implementation of the NetworkMultiplexer struct
type NetworkMultiplexer struct {
	in    map[string][]chan common.NetworkObject
	out   map[string][]chan chan common.NetworkObject
	Hosts []chan common.NetworkObject
}

func (m *NetworkMultiplexer) listen() {

	/*
		for {
			select {
			case c := m.out:
				for _, j := range m.out {
					for _, k := range j {
						log.Println("MULTI: Host added")
						m.Hosts = append(m.Hosts, <-k)
					}
				}
			case o := m.in:
				for _, j := range m.in {
					for _, k := range j {
						log.Println("MULTI: Transaction ", k, "to be sent to", len(m.Hosts))
						for _, l := range m.Hosts {
							go func(l chan common.NetworkObject) {
								l <- (k)
								log.Println("MULTI: Transaction sent to host")
							}(l)
						}
						log.Println("MULTI: Finished Processing")
					}
				}
			}
			log.Println("MULTI: Cycling")
		}
	*/
}

func (m *NetworkMultiplexer) AddListener(SwarmId string, c chan common.NetworkObject) {
	var new_out chan chan common.NetworkObject
	new_out <- c
	m.out[SwarmId] = append(m.out[SwarmId], new_out)
}

func (m *NetworkMultiplexer) SendNetworkObject(o common.NetworkObject) {
	var new_in chan common.NetworkObject
	new_in <- o
	m.in[o.SwarmId] = append(m.in[o.SwarmId], new_in)
}

//Listens to a stored chan NetworkObject (TCP)
func (m *NetworkMultiplexer) Listen(addr string) {
	panic("Not Implemented")
}

//Connects a to chan NetworkObject
func (m *NetworkMultiplexer) Connect(addr string) {
	//Possibly use the go.dial() method that is available
	panic("Not Implements")
}
