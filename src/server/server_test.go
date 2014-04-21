package main

import (
	"network"
	"quorum"
	"testing"
)

func TestNetworkedQuorum(t *testing.T) {
	// create a tcp server and 2 states
	// ms == messageSender
	ms, err := network.NewTCPServer(9988)
	if err != nil {
		println("fail")
	}
	// mh == messageHandler
	mh0, err := quorum.CreateState(ms, 0)
	if err != nil {
		println("fail")
	}
	mh1, err := quorum.CreateState(ms, 1)
	if err != nil {
		println("fail")
	}

	// add the states to each other
	mh0.AddParticipant(mh1.Self(), 1)
	mh1.AddParticipant(mh0.Self(), 0)

	// see if they initialize without issues
	mh0.Start()
	mh1.Start()
}
