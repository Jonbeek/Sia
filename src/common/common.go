package common

const (
	// How many bytes of entropy must be produced each entropy cycle
	ENTROPYVOLUME int = 32

	// How big a single slice of data is for a host, in bytes
	MINSLICESIZE int = 512
	MAXSLICESIZE int = 1048576 // 1 MB

	// How many hosts participate in each quorum
	// This number is chosen to minimize the probability of a single quorum
	// 	becoming more than 80% compromised by an attacker controlling 1/2 of
	// 	the network.
	QUORUMSIZE int = 192
)

type Message struct {
	Destination string
	Payload     string
}

type MessageHandler interface {
	HandleMessage(m Message)
}

type NetworkMultiplexer interface {
	AddListener(Swarmid string, c MessageHandler)
	SendMessage(o Message)
	Listen(addr string)
	Connect(addr string)
}
