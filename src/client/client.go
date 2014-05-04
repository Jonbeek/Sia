package main

import (
	"common"
	"common/crypto"
	"common/erasure"
	"fmt"
	"network"
)

// uploadSector splits a Sector into erasure-coded segments and distributes them across a quorum.
func uploadSector(sec *common.Sector, k int, quorum [common.QuorumSize]common.Address) (ring *common.Ring, err error) {
	// create erasure-coded segments
	ring, err = erasure.EncodeRing(sec, k)
	if err != nil {
		return
	}

	// for now we just send segment i to node i
	// this may need to be randomized for security
	for i := range quorum {
		m := &common.RPCMessage{
			quorum[i],
			"ServerHandler.UploadSegment",
			ring.Segs[i],
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
func downloadSector(hash crypto.Hash, ring *common.Ring, quorum [common.QuorumSize]common.Address) (sec *common.Sector, err error) {
	// send requests to each member of the quorum
	for i := range quorum {
		var seg common.Segment
		m := &common.RPCMessage{
			quorum[i],
			"ServerHandler.DownloadSegment",
			hash,
			&seg,
		}
		if network.SendRPCMessage(m) == nil {
			ring.AddSegment(&seg)
		}
	}

	// rebuild file
	sec, err = erasure.RebuildSector(ring)
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
