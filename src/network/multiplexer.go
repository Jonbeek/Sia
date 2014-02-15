package network

type NetworkObject struct {
	SwarmId       string
	TransactionId string
	BlockId       string
	Payload       string
}

type NetworkMultiplexer struct {
	Listeners       map[string][]chan NetworkObject
	Connectors      map[string][]chan NetworkObject
	Network_Objects []string
}

func (m *NetworkMultiplexer) AddListener(SwarmId string, c chan NetworkObject) {

}

func (m *NetworkMultiplexer) SendUpdate( /*t Transaction,*/ Dests []string) {

}

func (m *NetworkMultiplexer) Listen(addr string) {

}

func (m *NetworkMultiplexer) Connect(addr string) {

}
