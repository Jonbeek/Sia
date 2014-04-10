// Common contains structs, data, and interfaces that
// needs to be referenced by many packages but doesn't
// necessarily have an obvious place it belongs
package common

import (
	"time"
)

const (
	// How many bytes of entropy must be produced each entropy cycle
	EntropyVolume int = 32

	// How big a single slice of data is for a host, in bytes
	MinSliceSize int = 512
	MaxSliceSize int = 1048576 // 1 MB

	// How many hosts participate in each quorum
	// This number is chosen to minimize the probability of a single quorum
	// 	becoming more than 80% compromised by an attacker controlling 1/2 of
	// 	the network.
	QuorumSize int = 128

	// How long a single step in the consensus algorithm takes
	StepDuration time.Duration = 15 * time.Second
)

type Entropy [EntropyVolume]byte

// Messages are for sending things over the network.
// Each message has a single destination, and it is
// the job of the network package to interpret the
// destinations.
type Message struct {
	Destination []string // may not remain a string
	Payload     string
}

type MessageHandler interface {
	HandleMessage(m Message)
}
