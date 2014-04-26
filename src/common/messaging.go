package common

import (
	"encoding/binary"
	"fmt"
)

// An Identifier uniquely identifies a participant on a host.
type Identifier byte

// An Address couples an Identifier with its network address.
type Address struct {
	Id   Identifier
	Host string
	Port int
}

// Messages are for sending things over the network.
// Each message has a single destination, and it is
// the job of the network package to route a message
// to its intended destination.
type Message struct {
	Destination Address
	Payload     []byte
}

// A MessageHandler is a function that processes a message payload.
// A MessageHandler has an Address associated with it that is determined by a MessageRouter.
type MessageHandler interface {
	SetAddress(*Address)
	HandleMessage([]byte)
}

// A MessageRouter both transmits outgoing messages and processes incoming messages.
// It allows MessageHandlers to be associated with a given Identifier.
type MessageRouter interface {
	Address() Address
	AddMessageHandler(MessageHandler) Address
	SendMessage(*Message) error
}

// Turns an Identifier into a []byte
func (i Identifier) Marshal() (mi []byte) {
	mi = make([]byte, 1)
	mi[0] = byte(i)
	return
}

// Turns a []byte into an Identifier
func UnmarshalIdentifier(mi []byte) (i Identifier, err error) {
	if len(mi) != 1 {
		err = fmt.Errorf("Marshalled Identifier must be of length 1")
		return
	}

	i = Identifier(mi[0])
	return
}

// Turns an Address into a []byte
func (a *Address) Marshal() (ma []byte) {
	idAndHost := append(a.Id.Marshal(), []byte(a.Host)...)
	marshalledInt := make([]byte, 4)
	binary.PutUvarint(marshalledInt, uint64(a.Port))
	ma = append(idAndHost, marshalledInt...)
	return
}

// Turns a []byte into an Address
func UnmarshalAddress(ma []byte) (a *Address, err error) {
	if len(ma) < 5 {
		err = fmt.Errorf("marshalledAddress of insufficient length")
		return
	}

	a = new(Address)
	a.Id, err = UnmarshalIdentifier(ma[0:1])
	if err != nil {
		return
	}
	a.Host = string(ma[1 : len(ma)-4])
	port, _ := binary.Uvarint(ma[len(ma)-4:])
	a.Port = int(port)
	return
}
