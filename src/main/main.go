package main

import (
	//"common"
	"network"
	"quorum"
	"time"
)

func main() {
	println("starting")

	// create a networking server that can pass messages between hosts
	tcp, err := network.NewTCPServer(9988)
	if err != nil {
		println(err)
		return
	}

	// create states that communicate over the tcp server
	s4, _ := quorum.CreateState(tcp, 4)
	s3, _ := quorum.CreateState(tcp, 3)

	// add states to networking server
	tcp.MessageHandlers[4] = &s4
	tcp.MessageHandlers[3] = &s3

	// add states to each other
	s4.AddParticipant(s3.Self(), 3)
	s3.AddParticipant(s4.Self(), 4)

	// start the algorithms for each state
	s4.Start()
	//s3.Start()

	time.Sleep(time.Second)

	println("done")
}
