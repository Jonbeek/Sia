package main

import (
	"fmt"
	"network"
	"quorum"
)

func establishQuorum() {
	networkServer, err := network.NewTCPServer(9988)
	if err != nil {
		println(err)
		return
	}
	participant, err := quorum.CreateState(networkServer, 0)
	if err != nil {
		println(err)
		return
	}
	participant.Start()

	select {}
}

func main() {
	var input string
	for {
		print("Please enter a command:")
		fmt.Scanln(&input)

		switch input {
		default:
			println("unrecognized command")
		case "e":
			println("establishing new quorum")
			establishQuorum()
		case "q":
			return
		}
	}
}
