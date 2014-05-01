package main

import (
	"bytes"
	"common"
	"common/crypto"
	"common/erasure"
	"fmt"
	"io"
	"os"
)

// MessageHandlers

// uploadHandler is a MessageHandler that stores the received data in a byte slice.
// It uses a channel to signal that it has finished.
type uploadHandler struct {
	index byte
	data  []byte
	done  chan bool
}

func (tuh *uploadHandler) SetAddress(addr *common.Address) {
	return
}

func (uh *uploadHandler) HandleMessage(payload []byte) {
	uh.index = payload[0]
	uh.data = payload[1:]
	uh.done <- true
}

// downloadHandler is a MessageHandler that reconstructs a file from a set of segments.
// It uses a channel to signal that it has finished.
type downloadHandler struct {
	data     []byte
	segments []string
	indices  []uint8
	k, b     int
	done     chan bool
}

func (dh *downloadHandler) SetAddress(addr *common.Address) {
	return
}

func (dh *downloadHandler) HandleMessage(payload []byte) {
	// first byte is the segment index
	dh.indices = append(dh.indices, uint8(payload[0]))
	dh.segments = append(dh.segments, string(payload[1:]))
	// if enough segments have been collected, reconstruct the data
	if len(dh.segments) == dh.k {
		dh.data, _ = erasure.RebuildSector(dh.k, dh.b, dh.segments, dh.indices)
		dh.done <- true
	}
}

// TestTCPDownloadFile tests the NewTCPServer and DownloadFile functions.
// NewTCPServer must properly initialize a TCP server.
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
		m := new(common.Message)
		m.Destination = quorum[i]
		m.Payload = append([]byte{byte(i)}, []byte(segments[i])...)
		err = mr.SendMessage(m)
		if err != nil {
			return
		}
	}

	return
}

// DownloadFile retrieves the erasure-coded segments corresponding to a given file from a quorum.
// It reconstructs the original file from the segments using erasure.RebuildSector().
func DownloadFile(mr common.MessageRouter, fileHash crypto.Hash, length int, k int, quorum [common.QuorumSize]common.Address) (fileData []byte, err error) {
	// spawn a separate thread for each segment
	for i := range quorum {
		go func() {
			// send request
			m := new(common.Message)
			m.Destination = quorum[i]
			m.Payload = []byte{0x01}
			m.Payload = append(m.Payload, fileHash[:]...)
			mr.SendMessage(m)
			// wait for response
		}()
	}
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
