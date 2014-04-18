package network

import (
	"common"
	"testing"
	"time"
)

// a simple message handler
type TestMsgHandler struct {
	id     common.Identifier
	result string
}

func (t *TestMsgHandler) Identifier() common.Identifier {
	return t.id
}

func (t *TestMsgHandler) HandleMessage(payload []byte) {
	t.result = string(payload)
}

// TestNetworkNode tests the InitServer and SendMessage functions.
// InitServer must properly initialize a TCP server.
// SendMessage must succesfully deliver a message.
func TestTCPServer(t *testing.T) {
	// create TCPServer and add a message handler
	tcp, err := NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}

	// create message handler and add it to the TCPServer
	tmh := new(TestMsgHandler)
	tmh.id = 1
	tcp.AddMessageHandler(tmh)

	// send a message to be echoed
	m := common.Message{common.Address{1, "localhost", 9988}, []byte("hello, world!")}
	err = tcp.SendMessage(&m)
	if err != nil {
		t.Fatal("Failed to send message:", err)
	}
	// allow time for message to be processed
	time.Sleep(10 * time.Millisecond)
	resp := tmh.result
	if resp == "" {
		t.Fatal("Bad response: expected \"hello, world!\", got \"\"")
	}
	if string(resp) != "hello, world!" {
		t.Fatal("Bad response: expected \"hello, world!\", got \"" + string(resp) + "\"")
	}

	// send a message that should not trigger a MessageHandler
	tmh.result = ""
	m = common.Message{common.Address{2, "localhost", 9988}, []byte("won't be seen")}
	err = tcp.SendMessage(&m)
	if err != nil {
		t.Fatal("Failed to send message:", err)
	}
	// allow time for message to be processed
	time.Sleep(10 * time.Millisecond)
	resp = tmh.result
	if resp != "" {
		t.Fatal("Bad response: expected \"\", got \"" + string(resp) + "\"")
	}
}
