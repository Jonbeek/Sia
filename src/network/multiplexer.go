package network

//==========================================\\

//Temporary Implementation
//Probably will be implemented in network.go

type NetworkObject struct {
	SwarmId       string
	TransactionId string
	BlockId       string
	Payload       string
}

//============================================\\

//Implementation of the NetworkMultiplexer struct with several variables
type NetworkMultiplexer struct {
	Listeners       map[string][]chan NetworkObject // A map that uses strings to store channels of listeners of this multiplexer
	Connectors      map[string][]chan chan NetworkObject // A map that uses strings to store channels that the multiplexer is connected to
	Hosts 		[]chan NetworkObject            // A slice that contains the string IDs for all the local NetworkObjects
}

//Will add a new string ID and chan NetworkObject to Listeners
func (m *NetworkMultiplexer) AddListener(SwarmId string, c chan NetworkObject) {
	m.Listeners[SwarmId] = append(m.Listeners[SwarmId], c)
}

//Sends a Transaction t to every ID in Dests if the
func (m *NetworkMultiplexer) SendNetworkObject(o: NetworkObject) {

}

//Listens to a stored chan NetworkObject (TCP)
func (m *NetworkMultiplexer) Listen(addr string) {

}

//Connects a to chan NetworkObject
func (m *NetworkMultiplexer) Connect(addr string) {
	//Use the go.dial() method that is available
}
