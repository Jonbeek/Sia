package network

import (
	"common"
	"common/log"
	"encoding/json"
	"net"
)

type NetworkMultiplexer struct {
	in            chan common.NetworkMessage
	out           chan idHandler
	hosts         map[string][]common.NetworkMessageHandler
	createdServer bool
}

type idHandler struct {
	handler common.NetworkMessageHandler
	SwarmId string
}

func NewNetworkMultiplexer() common.NetworkMultiplexer {
	m := new(NetworkMultiplexer)
	m.in = make(chan common.NetworkMessage)
	m.out = make(chan idHandler)
	m.hosts = make(map[string][]common.NetworkMessageHandler)
	m.createdServer = false
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
				go s.HandleNetworkMessage(o)
			}
		}
	}
}

func (m *NetworkMultiplexer) AddListener(SwarmId string, c common.NetworkMessageHandler) {
	m.out <- idHandler{c, SwarmId}
}

func (m *NetworkMultiplexer) SendNetworkMessage(o common.NetworkMessage) {
	m.in <- o
}

func (m *NetworkMultiplexer) Listen(addr string) {

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Debugln("MULTI: ERROR COULD NOT CREATE SERVER WITH ADDRESS:", addr)
		log.Fatal("MULTI ERROR MESSAGE:", err)
	}

	log.Debugln("MULTI: SUCCESSFULLY CREATED ADDRESS: addr")
	log.Debugln("MULTI: LISTENING FOR CONNECTORS")

	for {
		conn, err := ln.Accept()
		if err != nil {
			// Or it might be conn.LocalAddr().String(), unsure which to use
			log.Debugln("MULTI: ERROR CONNECTION REFUSED FOR ADDRESS:", conn.RemoteAddr().String())
			log.Debugln("MULTI ERROR MESSAGE:", err)
			continue
		}
		log.Debugln("MULTI: CONNECTED TO ADDRESS: ", conn.RemoteAddr().String())
		log.Debugln("MULTI: SENDING MESSAGE TO ADDRESS: ", conn.RemoteAddr().String())

		msg := "TESTING CONNECTION WITH " + addr
		en := json.NewEncoder(conn)
		err = en.Encode(msg)
		if err != nil {
			log.Debugln(err)
			continue
		}

		de := json.NewDecoder(conn)
		err = de.Decode(msg)
		if err != nil {
			log.Debugln(err)
		}

		log.Debugln("MULTI: MESSAGE RECEIVED FROM:", conn.RemoteAddr().String())
		log.Debugln(msg)

		defer conn.Close()
	}

	defer ln.Close()

}

func (m *NetworkMultiplexer) Connect(addr string) {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Debugln("MULTI: ERROR CANNOT CONNECT TO SPECIFIED ADDRESS")
		log.Fatal("MULTI ERROR MESSAGE:", err)
	}

	log.Debugln("MULTI: CONNECTED TO ADDRESS:", addr)

	var msg string

	de := json.NewDecoder(conn)
	err = de.Decode(msg)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugln("MULTI: MESSAGE RECEIVED FROM:", addr)
	log.Debugln(msg)

	msg = "MESSAGE CONFIRMED AS RECEIVED"

	en := json.NewEncoder(conn)
	err = en.Encode(msg)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
}
