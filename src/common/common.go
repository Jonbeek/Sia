// Common contains structs, data, and interfaces that
// need to be referenced by other packages
package common

import (
	"time"
)

const (
	// How many bytes of entropy must be produced each entropy cycle
	EntropyVolume int = 32

	// How big a single segment of data is for a host, in bytes
	MinSegmentSize int = 512
	MaxSegmentSize int = 1048576 // 1 MB

	// How many participants are in each quorum
	// This number is chosen to minimize the probability of a single quorum
	// 	becoming more than 80% compromised by an attacker controlling 1/2 of
	// 	the network.
	QuorumSize int = 128

	// How long a single step in the consensus algorithm takes
	StepDuration time.Duration = 15 * time.Second
)

type Entropy [EntropyVolume]byte

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
