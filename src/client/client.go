package main

import (
	"common"
	"common/data"
	"fmt"
	"network"
)

// uploadSector splits a Sector into erasure-coded segments and distributes them across a quorum.
func uploadSector(sec *data.Sector, quorum [common.QuorumSize]common.Address) (err error) {
	// create erasure-coded segments
	segments, err := sec.Encode()
	if err != nil {
		return
	}

	// for now we just send segment i to node i
	// this may need to be randomized for security
	for i := range quorum {
		m := &common.RPCMessage{
			quorum[i],
			"ServerHandler.UploadSegment",
			segments[i],
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
func downloadSector(sec *data.Sector, quorum [common.QuorumSize]common.Address) error {
	// send requests to each member of the quorum
	numSent := 0
	var segments []string
	var indices []uint8
	for i := range quorum {
		var seg data.Segment
		m := &common.RPCMessage{
			quorum[i],
			"ServerHandler.DownloadSegment",
			sec.Hash,
			&seg,
		}
		if network.SendRPCMessage(m) == nil {
			segments = append(segments, seg.Data)
			indices = append(indices, seg.Index)
			numSent++
		}
	}
	if numSent < sec.GetRedundancy() {
		return fmt.Errorf("too few hosts reachable: needed %v, reached %v", sec.GetRedundancy(), numSent)
	}

	// rebuild file
	return sec.Rebuild(segments, indices)
}

func joinQuorum(mr common.MessageRouter) (q [common.QuorumSize]common.Address) {
	// for a := range q {

	// }
	return q
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
