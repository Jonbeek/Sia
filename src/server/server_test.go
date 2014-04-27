package main

import (
	"network"
	"quorum"
	"testing"
)

func TestNetworkedQuorum(t *testing.T) {
	// create a tcp server and 2 states
	// ms == messageSender
	ms, err := network.NewTCPServer(9980)
	if err != nil {
		println("fail")
	}
	// mh == messageHandler
	_, err = quorum.CreateState(ms)
	if err != nil {
		println("fail")
	}
	_, err = quorum.CreateState(ms)
	if err != nil {
		println("fail")
	}

	// more code here
}
