package main

import (
	"bytes"
	"common"
	"common/crypto"
	"common/erasure"
	"io/ioutil"
	"network"
	"os"
	"testing"
)

// serverHandler is a MessageHandler that processes messages sent to a server.
// It uses a channel to signal that it has finished processing.
type serverHandler struct {
	mr    common.MessageRouter
	index byte
	data  []byte
	done  chan bool
}

func (sh *serverHandler) SetAddress(addr *common.Address) {
	return
}

func (sh *serverHandler) HandleMessage(payload []byte) {
	// first byte contains the message type
	switch payload[0] {
	case 0: // store segment
		sh.index = payload[1]
		sh.data = payload[2:]
	case 1: // reply to sender with segment
		var err error
		m := new(common.Message)
		m.Destination, err = common.UnmarshalAddress(payload[1:])
		if err != nil {
			return
		}
		m.Payload = append([]byte{0x00, sh.index}, sh.data...)
		sh.mr.SendMessage(m)
	}
	sh.done <- true
}

// TestTCPUploadFile tests the NewTCPServer and UploadFile functions.
// NewTCPServer must properly initialize a TCP server.
// UploadFile must succesfully distribute a file among a quorum.
// The uploaded file must be successfully reconstructed.
func TestTCPUploadFile(t *testing.T) {
	// create TCPServer
	tcp, err := network.NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer tcp.Close()

	// create quorum
	var q [common.QuorumSize]common.Address
	var shs [common.QuorumSize]serverHandler
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qtcp, err := network.NewTCPServer(9000 + i)
		defer qtcp.Close()
		if err != nil {
			t.Fatal("Failed to initialize TCPServer:", err)
		}
		shs[i].done = make(chan bool, 1)
		q[i].Id = qtcp.AddMessageHandler(&shs[i]).Id
	}

	// create file
	file, err := os.Create("InputFile")
	if err != nil {
		t.Fatal("Failed to create file \"InputFile\"")
	}
	defer file.Close()
	defer os.Remove("InputFile")

	fileData, err := crypto.RandomByteSlice(70000)
	if err != nil {
		t.Fatal("Could not generate test data:", err)
	}

	err = ioutil.WriteFile("InputFile", fileData, 0644)
	if err != nil {
		t.Fatal("Failed to write to file InputFile:", err)
	}

	// calculate hash
	origHash, err := crypto.CalculateHash(fileData)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	// upload file to quorum
	k := common.QuorumSize / 2
	b, err := UploadFile(tcp, file, k, q)
	if err != nil {
		t.Fatal("Failed to upload file:", err)
	}

	// wait for all participants to complete
	for i := range shs {
		<-shs[i].done
	}

	// rebuild file from first k segments
	segments := make([]string, k)
	indices := make([]uint8, k)
	for i := 0; i < k; i++ {
		segments[i] = string(shs[i].data)
		indices[i] = uint8(shs[i].index)
	}

	rebuiltData, err := erasure.RebuildSector(k, b, segments, indices)
	if err != nil {
		t.Fatal("Failed to rebuild file:", err)
	}
	// remove padding
	rebuiltData = rebuiltData[:len(fileData)]

	// check hash
	rebuiltHash, err := crypto.CalculateHash(rebuiltData)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	if origHash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}

// NewTCPServer must properly initialize a TCP server.
// DownloadFile must successfully retrieve a file from a quorum.
// The downloaded file must match the original file.
func TestTCPDownloadFile(t *testing.T) {
	// create file
	fileData, err := crypto.RandomByteSlice(70000)
	if err != nil {
		t.Fatal("Could not generate test data:", err)
	}

	// calculate hash
	origHash, err := crypto.CalculateHash(fileData)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	// encode file
	k := common.QuorumSize / 2
	bytesPerSegment := len(fileData) / k
	if bytesPerSegment%64 != 0 {
		bytesPerSegment += 64 - (bytesPerSegment % 64)
		padding := k*bytesPerSegment - len(fileData)
		fileData = append(fileData, bytes.Repeat([]byte{0x00}, padding)...)
	}
	segments, err := erasure.EncodeRing(k, bytesPerSegment, fileData)
	if err != nil {
		t.Fatal("Failed to encode file data:", err)
	}

	// create TCPServer
	tcp, err := network.NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer tcp.Close()
	ch := new(clientHandler)
	ch.k, ch.b = k, bytesPerSegment
	ch.done = make(chan bool, 1)
	tcp.AddMessageHandler(ch)

	// create quorum
	var q [common.QuorumSize]common.Address
	var shs [common.QuorumSize]serverHandler
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qtcp, err := network.NewTCPServer(9000 + i)
		if err != nil {
			t.Fatal("Failed to initialize TCPServer:", err)
		}
		shs[i].mr = qtcp
		shs[i].index = byte(i)
		shs[i].data = []byte(segments[i])
		shs[i].done = make(chan bool, 1)
		q[i].Id = qtcp.AddMessageHandler(&shs[i]).Id
	}

	// download file from quorum
	downData, err := DownloadFile(tcp, ch, origHash, 70000, k, q)
	if err != nil {
		t.Fatal("Failed to download file:", err)
	}

	// check hash
	rebuiltHash, err := crypto.CalculateHash(downData)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	if origHash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}
