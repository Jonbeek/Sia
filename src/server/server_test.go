package main

import (
	"network"
	"quorum"
	"testing"
)

func TestNetworkedQuorum(t *testing.T) {
	// create a MessageRouter and 2 states
	rpcs, err := network.NewRPCServer(9980)
	if err != nil {
		println("fail")
	}

	_, err = quorum.CreateState(rpcs)
	if err != nil {
		println("fail")
	}
	_, err = quorum.CreateState(rpcs)
	if err != nil {
		println("fail")
	}

	// more code here
}
