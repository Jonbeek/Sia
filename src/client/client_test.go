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
	var uhs [common.QuorumSize]uploadHandler
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qtcp, err := network.NewTCPServer(9000 + i)
		defer qtcp.Close()
		if err != nil {
			t.Fatal("Failed to initialize TCPServer:", err)
		}
		uhs[i].done = make(chan bool, 1)
		q[i].Id = qtcp.AddMessageHandler(&uhs[i]).Id
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
	k := 50
	b, err := UploadFile(tcp, file, k, q)
	if err != nil {
		t.Fatal("Failed to upload file:", err)
	}

	// wait for all participants to complete
	for i := range uhs {
		<-uhs[i].done
	}

	// rebuild file from first k segments
	segments := make([]string, k)
	indices := make([]uint8, k)
	for i := 0; i < k; i++ {
		segments[i] = string(uhs[i].data)
		indices[i] = uint8(uhs[i].index)
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

// a simple message handler
// sends data to dest
// uses a channel to signal when handler has been called
type TestFileHandler struct {
	tcpServ *network.TCPServer
	dest    common.Address
	data    string
	done    chan bool
}

func (tfh *TestFileHandler) SetAddress(addr *common.Address) {
	return
}

func (tfh *TestFileHandler) HandleMessage(payload []byte) {
	m := new(common.Message)
	m.Destination = tfh.dest
	m.Payload = []byte(tfh.data)
	tfh.tcpServ.SendMessage(m)

	tfh.done <- true
}

// NewTCPServer must properly initialize a TCP server.
// DownloadFile must successfully retrieve a file from a quorum.
// The downloaded file must match the original file.
func TestTCPDownloadFile(t *testing.T) {
	t.Skip()
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
	k := 50
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
	tdh := new(downloadHandler)
	tdh.segments = make([]string, k)
	tdh.indices = make([]uint8, k)
	tdh.k, tdh.b = k, bytesPerSegment
	tcp.AddMessageHandler(tdh)

	// create quorum
	var q [common.QuorumSize]common.Address
	var tfhs [common.QuorumSize]TestFileHandler
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qtcp, err := network.NewTCPServer(9000 + i)
		if err != nil {
			t.Fatal("Failed to initialize TCPServer:", err)
		}
		tfhs[i].tcpServ = qtcp
		tfhs[i].dest = tcp.Address()
		tfhs[i].data = segments[i]
		tfhs[i].done = make(chan bool, 1)
		q[i].Id = qtcp.AddMessageHandler(&tfhs[i]).Id
	}

	// download file from quorum
	downData, err := DownloadFile(tcp, origHash, len(fileData), k, q)
	if err != nil {
		t.Fatal("Failed to download file:", err)
	}

	// wait for download to complete
	<-tdh.done

	// check hash
	rebuiltHash, err := crypto.CalculateHash(downData)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	if origHash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}
