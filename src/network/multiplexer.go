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
	handler common.MessageHandler
	SwarmId string
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
			m.hosts[c.SwarmId] = append(m.hosts[c.SwarmId], c.handler)
		case o := <-m.in:
			log.Debugln("MULTI: Transaction ", o, "to be sent to", len(m.hosts[o.SwarmId]))
			for _, s := range m.hosts[o.SwarmId] {
				go s.HandleMessage(o)
			}
		}
	}
}

func (m *NetworkMultiplexer) AddListener(SwarmId string, c common.MessageHandler) {
	m.out <- idHandler{c, SwarmId}
}

func (m *NetworkMultiplexer) SendMessage(o common.Message) {
	m.in <- o
}

func (m *NetworkMultiplexer) Listen(addr string) {

}

func (m *NetworkMultiplexer) Connect(addr string) {

}
