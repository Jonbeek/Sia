package common

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
