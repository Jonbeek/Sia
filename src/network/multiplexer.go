package network

import (
	"common"
	"log"
)

type NetworkMultiplexer struct {
	in    chan common.NetworkMessage
	out   chan Network_Message
	Hosts map[string][]common.NetworkMessageHandler
}

type Network_Message struct {
	handler common.NetworkMessageHandler
	SwarmId string
}

func NewNetworkMultiplexer() *NetworkMultiplexer {
	m := new(NetworkMultiplexer)
	m.in = make(chan common.NetworkMessage)
	m.out = make(chan Network_Message)
	m.Hosts = make(map[string][]common.NetworkMessageHandler)
	go m.listen()
	return m
}

func (m *NetworkMultiplexer) listen() {
	for {
		select {
		case c := <-m.out:
			m.Hosts[c.SwarmId] = append(m.Hosts[c.SwarmId], c.handler)
		case o := <-m.in:
			log.Println("MULTI: Transaction ", o, "to be sent to", len(m.Hosts[o.SwarmId]))
			for _, s := range m.Hosts[o.SwarmId] {
				go s.HandleNetworkMessage(o)
			}
		}
	}
}

func (m *NetworkMultiplexer) AddListener(SwarmId string, c common.NetworkMessageHandler) {
	m.out <- Network_Message{c, SwarmId}
}

func (m *NetworkMultiplexer) SendNetworkMessage(o common.NetworkMessage) {
	m.in <- o
}

func (m *NetworkMultiplexer) Listen(addr string) {

}

func (m *NetworkMultiplexer) Connect(addr string) {

}
