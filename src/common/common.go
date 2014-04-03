package common

import (
	"time"
)

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

type Update interface {
	SwarmId() string
	UpdateId() string
	MarshalString() string
	Type() string
}

type Message struct {
	SwarmId  string
	UpdateId string
	Payload  string
	Type     string
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

// swarmsize is the number of hosts managing each set of files
var SWARMSIZE int = 192

var STATEINFORMEDDELTA time.Duration = 1 * time.Second
