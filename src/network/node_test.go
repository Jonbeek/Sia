package network

import (
	"bytes"
	"common"
	"common/crypto"
	"common/erasure"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// a simple message handler
// stores the received message in the result field
// uses a channel to signal when handler has been called
type TestMsgHandler struct {
	result string
	done   chan bool
}

func (tmh *TestMsgHandler) SetAddress(addr *common.Address) {
	return
}

func (tmh *TestMsgHandler) HandleMessage(payload []byte) {
	tmh.result = string(payload)
	tmh.done <- true
}

// TestTCPSendMessage tests the NewTCPServer and SendMessage functions.
// NewTCPServer must properly initialize a TCP server.
// SendMessage must succesfully deliver a message.
func TestTCPSendMessage(t *testing.T) {
	// create TCPServer and add a message handler
	tcp, err := NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer tcp.Close()

	// create message handler and add it to the TCPServer
	tmh := new(TestMsgHandler)
	tmh.done = make(chan bool, 1)
	tcp.AddMessageHandler(tmh)

	// send a message to be echoed
	m := common.Message{common.Address{1, "localhost", 9988}, []byte("hello, world!")}
	err = tcp.SendMessage(&m)
	if err != nil {
		t.Fatal("Failed to send message:", err)
	}

	// wait for handler to be triggered
	<-tmh.done

	if tmh.result == "" {
		t.Fatal("Bad response: expected \"hello, world!\", got \"\"")
	}
	if tmh.result != "hello, world!" {
		t.Fatal("Bad response: expected \"hello, world!\", got \"" + tmh.result + "\"")
	}

	// send a message that should not trigger a MessageHandler
	tmh.result = ""
	m = common.Message{common.Address{2, "localhost", 9988}, []byte("won't be seen")}
	err = tcp.SendMessage(&m)
	if err != nil {
		t.Fatal("Failed to send message:", err)
	}

	// arbitrary wait, because <-tmh.done will block
	time.Sleep(10 * time.Millisecond)

	if tmh.result != "" {
		t.Fatal("Bad response: expected \"\", got \"" + tmh.result + "\"")
	}

	// send a message longer than 1024 bytes
	tmh.result = ""
	m = common.Message{common.Address{1, "localhost", 9988}, bytes.Repeat([]byte("b"), 9001)}
	err = tcp.SendMessage(&m)
	if err != nil {
		t.Fatal("Failed to send message:", err)
	}

	// wait for handler to be triggered
	<-tmh.done

	if len(tmh.result) != 9001 {
		t.Fatal("Bad response: expected 9001 bytes, got", len(tmh.result))
	}
}

// a simple message handler
// writes the received data to a file handle
// uses a channel to signal when handler has been called
type TestFileHandler struct {
	file *os.File
	done chan bool
}

func (tfh *TestFileHandler) SetAddress(addr *common.Address) {
	return
}

func (tfh *TestFileHandler) HandleMessage(payload []byte) {
	tfh.file.Write(payload)
	tfh.done <- true
}

// TestTCPSendSegment tests the NewTCPServer and SendSegment functions.
// NewTCPServer must properly initialize a TCP server.
// SendSegment must succesfully transfer a segment.
func TestTCPSendSegment(t *testing.T) {
	// create TCPServer and add a message handler
	tcp, err := NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer tcp.Close()

	// create message handler and add it to the TCPServer
	tfh := new(TestFileHandler)
	tfh.done = make(chan bool, 1)
	tfh.file, err = os.Create("OutputFile")
	if err != nil {
		t.Fatal("Failed to create file \"OutputFile\"")
	}
	defer tfh.file.Close()
	defer os.Remove("OutputFile")
	tcp.AddMessageHandler(tfh)

	// create file
	file, err := os.Create("InputFile")
	if err != nil {
		t.Fatal("Failed to create file \"InputFile\"")
	}
	defer file.Close()
	defer os.Remove("InputFile")
	err = ioutil.WriteFile("InputFile", bytes.Repeat([]byte("b"), 9001), 0644)
	if err != nil {
		t.Fatal("Failed to write to file InputFile")
	}

	// send file
	dest := common.Address{1, "localhost", 9988}
	err = tcp.SendSegment(file, &dest)
	if err != nil {
		t.Fatal("Failed to send file:", err)
	}

	// wait for handler to be triggered
	<-tfh.done

	in, out := make([]byte, 9001), make([]byte, 9001)
	io.ReadFull(file, in)
	io.ReadFull(tfh.file, out)
	if bytes.Compare(in, out) != 0 {
		t.Fatal("Files do not match")
	}
}

// a simple message handler
// stores the received data in a byte slice
// uses a channel to signal when handler has been called
type TestUploadHandler struct {
	index byte
	data  []byte
	done  chan bool
}

func (tuh *TestUploadHandler) SetAddress(addr *common.Address) {
	return
}

func (tuh *TestUploadHandler) HandleMessage(payload []byte) {
	tuh.index = payload[0]
	tuh.data = payload[1:]
	tuh.done <- true
}

// TestTCPUploadFile tests the NewTCPServer and UploadFile functions.
// NewTCPServer must properly initialize a TCP server.
// UploadFile must succesfully distribute a file among a quorum.
// The uploaded file must be successfully reconstructed.
func TestTCPUploadFile(t *testing.T) {
	// create TCPServer
	tcp, err := NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer tcp.Close()

	// create quorum
	var q [common.QuorumSize]common.Address
	var tuhs [common.QuorumSize]TestUploadHandler
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qtcp, err := NewTCPServer(9000 + i)
		if err != nil {
			t.Fatal("Failed to initialize TCPServer:", err)
		}
		tuhs[i].done = make(chan bool, 1)
		q[i].Id = qtcp.AddMessageHandler(&tuhs[i]).Id
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
	b, err := tcp.UploadFile(file, k, q)
	if err != nil {
		t.Fatal("Failed to upload file:", err)
	}

	// wait for all participants to complete
	for i := range tuhs {
		<-tuhs[i].done
	}

	// rebuild file from first k segments
	segments := make([]string, k)
	indices := make([]uint8, k)
	for i := 0; i < k; i++ {
		segments[i] = string(tuhs[i].data)
		indices[i] = uint8(tuhs[i].index)
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
