package main

import (
	"common"
	"common/erasure"
	"fmt"
	"network"
)

// uploadSector splits a Sector into erasure-coded segments and distributes them across a quorum.
func uploadSector(sec *common.Sector, quorum [common.QuorumSize]common.Address) (err error) {
	// create erasure-coded segments
	ring, err := erasure.EncodeRing(sec)
	if err != nil {
		return
	}

	// for now we just send segment i to node i
	// this may need to be randomized for security
	for i := range quorum {
		m := &common.RPCMessage{
			quorum[i],
			"ServerHandler.UploadSegment",
			ring[i],
			nil,
		}
		err = network.SendRPCMessage(m)
		if err != nil {
			return
		}
	}

	return
}

// downloadSector retrieves the erasure-coded segments corresponding to a given Sector from a quorum.
// It reconstructs the original data from the segments and returns the complete Sector
func downloadSector(sec *common.Sector, quorum [common.QuorumSize]common.Address) (err error) {
	// send requests to each member of the quorum
	numSent := 0
	var segs []common.Segment
	for i := range quorum {
		var seg common.Segment
		m := &common.RPCMessage{
			quorum[i],
			"ServerHandler.DownloadSegment",
			sec.Hash,
			&seg,
		}
		if network.SendRPCMessage(m) == nil {
			segs = append(segs, seg)
			numSent++
		}
	}
	if numSent < sec.GetRedundancy() {
		return fmt.Errorf("too few hosts reachable: needed %v, reached %v", sec.GetRedundancy(), numSent)
	}

	// rebuild file
	err = erasure.RebuildSector(sec, segs[:sec.GetRedundancy()])
	return
}

func main() {
	var input string
	for {
		print("Please enter a command:")
		fmt.Scanln(&input)

		switch input {
		default:
			println("unrecognized command")
		case "j":
			println("joining quorum")
		case "u":
			println("uploading file")
		case "q":
			return
		}
	}
}
