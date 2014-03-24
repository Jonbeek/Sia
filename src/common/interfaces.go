package common

type Record interface {
	Type() string
	MarshalString() string
}

type Update interface {
	SwarmId() string
	UpdateId() string
	MarshalString() string
	Type() string
}

type NetworkMessage struct {
	SwarmId  string
	UpdateId string
	Payload  string
	Type     string
}

type NetworkMessageHandler interface {
	HandleNetworkMessage(m NetworkMessage)
}

type NetworkMultiplexer interface {
	AddListener(Swarmid string, c NetworkMessageHandler)
	SendNetworkMessage(o NetworkMessage)
	Listen(addr string)
	Connect(addr string)
}
