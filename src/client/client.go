package main

import (
	"bytes"
	"common"
	"common/crypto"
	"common/erasure"
	"encoding/gob"
	"fmt"
	"network"
)

// clientHandler is a MessageHandler that processes messages sent to a client.
// It uses a channel to signal that it has finished processing.
type clientHandler struct {
	segments []string
	indices  []uint8
	k, b     int
	done     chan struct{}
}

func (ch *clientHandler) SetAddress(addr *common.Address) {
	return
}

func (ch *clientHandler) HandleMessage(payload []byte) {
	// first byte contains the message type
	switch payload[0] {
	case 0:
		ch.indices = append(ch.indices, uint8(payload[1]))
		ch.segments = append(ch.segments, string(payload[2:]))
		if len(ch.segments) == ch.k {
			ch.done <- struct{}{}
		}
	}
}

// UploadFile splits a file into erasure-coded segments and distributes them across a quorum.
// k is the number of non-redundant segments.
// The file is padded to satisfy the erasure-coding requirements that:
//     len(fileData) = k*bytesPerSegment, and:
//     bytesPerSegment % 64 = 0
func UploadFile(mr common.MessageRouter, file *os.File, k int, quorum [common.QuorumSize]common.Address) (bytesPerSegment int, err error) {
	// read file
	fileInfo, err := file.Stat()
	if err != nil {
		return
	}
	if fileInfo.Size() > int64(common.QuorumSize*common.MaxSegmentSize) {
		err = fmt.Errorf("File exceeds maximum per-quorum size")
		return
	}
	fileData := make([]byte, fileInfo.Size())
	_, err = io.ReadFull(file, fileData)
	if err != nil {
		return
	}

	// calculate EncodeRing parameters, padding file if necessary
	bytesPerSegment = len(fileData) / k
	if bytesPerSegment%64 != 0 {
		bytesPerSegment += 64 - (bytesPerSegment % 64)
		padding := k*bytesPerSegment - len(fileData)
		fileData = append(fileData, bytes.Repeat([]byte{0x00}, padding)...)
	}

	// create erasure-coded segments
	segments, err := erasure.EncodeRing(k, bytesPerSegment, fileData)
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

// DownloadFile retrieves the erasure-coded segments corresponding to a given file from a quorum.
// It reconstructs the original file from the segments using erasure.RebuildSector().
func DownloadFile(mr common.MessageRouter, ch *clientHandler, fileHash crypto.Hash, length int, k int, quorum [common.QuorumSize]common.Address) (data []byte, err error) {
	// send requests to each member of the quorum
	numSent := 0
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
	if numSent < k {
		err = fmt.Errorf("too few hosts reachable: needed %v, reached %v", k, numSent)
		return
	}

	// wait for responses
	<-ch.done

	// rebuild file
	data, err = erasure.RebuildSector(ch.k, ch.b, ch.segments[:ch.k], ch.indices[:ch.k])
	// remove padding
	data = data[:length]

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
