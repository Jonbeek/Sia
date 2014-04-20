package network

import (
	"bytes"
	"common"
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
	id     common.Identifier
	result string
	done   chan bool
}

func (tmh *TestMsgHandler) Identifier() common.Identifier {
	return tmh.id
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
	tmh.id = 1
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
	id   common.Identifier
	file *os.File
	done chan bool
}

func (tfh *TestFileHandler) Identifier() common.Identifier {
	return tfh.id
}

func (tfh *TestFileHandler) HandleMessage(payload []byte) {
	tfh.file.Write(payload)
	tfh.done <- true
}

// TestTCPSendFile tests the NewTCPServer and SendFile functions.
// NewTCPServer must properly initialize a TCP server.
// SendFile must succesfully transfer a file.
func TestTCPSendFile(t *testing.T) {
	// create TCPServer and add a message handler
	tcp, err := NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer tcp.Close()

	// create message handler and add it to the TCPServer
	tfh := new(TestFileHandler)
	tfh.id = 1
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
	err = tcp.SendFile(file, &dest)
	if err != nil {
		t.Fatal("Failed to send message:", err)
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
