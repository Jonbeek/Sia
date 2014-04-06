package network

import (
	"common"
	"common/log"
)

type NetworkMultiplexer struct {
	in    chan common.Message
	out   chan idHandler
	hosts map[string][]common.MessageHandler
}

type idHandler struct {
	handler     common.MessageHandler
	Destination string
}

func NewNetworkMultiplexer() common.NetworkMultiplexer {
	m := new(NetworkMultiplexer)
	m.in = make(chan common.Message)
	m.out = make(chan idHandler)
	m.hosts = make(map[string][]common.MessageHandler)
	go m.listen()
	return m
}

func (m *NetworkMultiplexer) listen() {
	for {
		select {
		case c := <-m.out:
			m.hosts[c.Destination] = append(m.hosts[c.Destination], c.handler)
		case o := <-m.in:
			log.Debugln("MULTI: Transaction ", o, "to be sent to", len(m.hosts[o.Destination]))
			for _, s := range m.hosts[o.Destination] {
				go s.HandleMessage(o)
			}
		}
	}
}

func (m *NetworkMultiplexer) AddListener(Destination string, c common.MessageHandler) {
	m.out <- idHandler{c, Destination}
}

func (m *NetworkMultiplexer) SendMessage(o common.Message) {
	m.in <- o
}

func (m *NetworkMultiplexer) Listen(addr string) {

}

func (m *NetworkMultiplexer) Connect(addr string) {

}
