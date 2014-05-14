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
	router   common.MessageRouter
	SectorDB map[crypto.Hash]*common.RingHeader
)

// uploadSector splits a Sector into erasure-coded segments and distributes them across a quorum.
// It creates a Ring from its arguments and stores it in the SectorDB.
func uploadSector(sec *common.Sector) (err error) {
	// look up Sector in SectorDB
	rh := SectorDB[sec.Hash]
	if rh == nil {
		return fmt.Errorf("Sector not found in database")
	}

	// create ring
	ring, err := erasure.EncodeRing(sec, rh.Params)
	if err != nil {
		return
	}

	// calculate and store segment hashes
	for i := range ring {
		rh.SegHashes[i], err = crypto.CalculateHash(ring[i].Data)
		if err != nil {
			return
		}
	}

	// for now we just send segment i to host i
	// this may need to be randomized for security
	for i := range rh.Hosts {
		err = router.SendMessage(&common.Message{
			Dest: rh.Hosts[i],
			Proc: "Server.UploadSegment",
			Args: ring[i],
			Resp: nil,
		})
		if err != nil {
			return
		}
	}

	return
}

// downloadSector retrieves the erasure-coded segments corresponding to a given Sector from a quorum.
// It reconstructs the original data from the segments and returns the complete Sector
func downloadSector(hash crypto.Hash) (sec *common.Sector, err error) {
	// look up Sector in SectorDB
	rh := SectorDB[hash]
	if rh == nil {
		err = fmt.Errorf("Sector not found in database")
		return
	}

	// send requests to each member of the quorum
	var segs []common.Segment
	for i := range rh.Hosts {
		var seg common.Segment
		sendErr := router.SendMessage(&common.Message{
			Dest: rh.Hosts[i],
			Proc: "Server.DownloadSegment",
			Args: rh.SegHashes[i],
			Resp: &seg,
		})
		if sendErr == nil {
			segs = append(segs, seg)
		} else {
			fmt.Println(sendErr)
		}
	}

	// rebuild file
	sec, err = erasure.RebuildSector(segs, rh.Params)
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

func generateSector(q common.Quorum) (s *common.Sector, err error) {
	if q[0].Port == 0 {
		err = fmt.Errorf("you must connect to a quorum first")
		return
	}
	data, err := crypto.RandomByteSlice(70000)
	if err != nil {
		return
	}
	s, err = common.NewSector(data)
	if err != nil {
		return
	}
	SectorDB[s.Hash] = &common.RingHeader{
		Hosts:  q,
		Params: s.CalculateParams(common.QuorumSize / 2),
	}
	return
}

func main() {
	router, _ = network.NewRPCServer(9989)
	defer router.Close()
	SectorDB = make(map[crypto.Hash]*common.RingHeader)
	var (
		input string
		q     common.Quorum
		s     *common.Sector
		h     crypto.Hash
		err   error
	)
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
		case "g":
			fmt.Println("generating Sector")
			s, err = generateSector(q)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("failed to generate Sector")
				break
			}
			h = s.Hash
			fmt.Println("created Sector with hash", h[:10])
		case "u":
			fmt.Println("uploading file")
			err = uploadSector(s)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("upload failed")
				break
			}
			fmt.Println("upload successful")
		case "d":
			fmt.Println("downloading file")
			rs, err := downloadSector(h)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Println("download failed")
				break
			}
			rh, err := crypto.CalculateHash(rs.Data)
			if err != nil {
				fmt.Println("error:", err)
				break
			}
			fmt.Println("download successful")
			fmt.Println("hash:", rh[:10])
		case "q":
			return
		}
	}
}
