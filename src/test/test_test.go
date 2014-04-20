package test

import (
	"network"
	"quorum"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	// ms == messageSender
	ms, err := network.NewTCPServer(9988)
	if err != nil {
		t.Fatal(err)
	}

	// mh == messageHandler
	mh0, err := quorum.CreateState(ms, 0)
	if err != nil {
		t.Fatal(err)
	}

	mh1, err := quorum.CreateState(ms, 1)
	if err != nil {
		t.Fatal(err)
	}

	mh0.AddParticipant(mh1.Self(), 1)
	mh1.AddParticipant(mh0.Self(), 0)

	mh0.Start()
	mh1.Start()

	time.Sleep(time.Second)
}
