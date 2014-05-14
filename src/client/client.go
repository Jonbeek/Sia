package main

import (
	"common"
	"common/crypto"
	"common/erasure"
	"fmt"
	"network"
)

// global variables
// (with apologies to Haskell)
var (
	mr       common.MessageRouter
	SectorDB map[crypto.Hash]*common.Ring
)

// uploadSector splits a Sector into erasure-coded segments and distributes them across a quorum.
// It creates a Ring from its arguments and stores it in the SectorDB.
func uploadSector(sec *common.Sector, k int, q common.Quorum) (err error) {
	// create ring
	ring, segs, err := erasure.EncodeRing(sec, k)
	if err != nil {
		return
	}
	ring.Hosts = q

	// for now we just send segment i to node i
	// this may need to be randomized for security
	for i := range q {
		err = mr.SendMessage(&common.Message{
			Dest: q[i],
			Proc: "Server.UploadSegment",
			Args: segs[i],
			Resp: nil,
		})
		if err != nil {
			return
		}
	}

	// add ring to SectorDB
	SectorDB[sec.Hash] = ring

	return
}

// downloadSector retrieves the erasure-coded segments corresponding to a given Sector from a quorum.
// It reconstructs the original data from the segments and returns the complete Sector
func downloadSector(hash crypto.Hash) (sec *common.Sector, err error) {
	// retrieve ring from SectorDB
	ring := SectorDB[hash]
	if ring == nil {
		err = fmt.Errorf("hash not present in database")
		return
	}

	// send requests to each member of the quorum
	var segs []common.Segment
	for i := range ring.Hosts {
		var seg common.Segment
		sendErr := mr.SendMessage(&common.Message{
			Dest: ring.Hosts[i],
			Proc: "Server.DownloadSegment",
			Args: ring.SegHashes[i],
			Resp: &seg,
		})
		if sendErr == nil {
			segs = append(segs, seg)
		} else {
			fmt.Println(sendErr)
		}
	}

	// rebuild file
	sec, err = erasure.RebuildSector(ring, segs)
	return
}

func readQuorumAddresses() (q [common.QuorumSize]common.Address) {
	var input int
	for i := range q {
		fmt.Print("Please enter port number ", i, ": ")
		fmt.Scanln(&input)
		q[i] = common.Address{2, "localhost", input}
	}
	return
}

func main() {
	mr, _ = network.NewRPCServer(9989)
	defer mr.Close()
	SectorDB = make(map[crypto.Hash]*common.Ring)
	var (
		input string
		q     [common.QuorumSize]common.Address
		s, rs *common.Sector
		h, rh crypto.Hash
	)
	data, err := crypto.RandomByteSlice(70000)
	s, err = common.NewSector(data)
	h, err = crypto.CalculateHash(data)
	if err != nil {
		fmt.Println("error:", err)
	}
	for {
		fmt.Print("Please enter a command: ")
		fmt.Scanln(&input)

		switch input {
		default:
			fmt.Println("unrecognized command")
		case "j":
			fmt.Println("joining quorum")
			q = readQuorumAddresses()
			fmt.Println("connected to quorum")
		case "u":
			fmt.Println("uploading file")
			err = uploadSector(s, 2, q)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("upload failed")
				break
			}
			fmt.Println("upload successful")
			fmt.Println("hash:", h[:])
		case "d":
			fmt.Println("downloading file")
			rs, err = downloadSector(h)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("download failed")
				break
			}
			rh, err = crypto.CalculateHash(rs.Data)
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			fmt.Println("download successful")
			fmt.Println("hash:", rh[:])
		case "q":
			return
		}
	}
}
