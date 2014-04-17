package network

import (
	"common"
	"testing"
	"time"
)

// a simple message handler
type TestMsgHandler struct {
	result string
}

func (t *TestMsgHandler) HandleMessage(payload []byte) {
	t.result = string(payload)
}

// TestNetworkNode tests the InitServer and SendMessage functions.
// InitServer must properly initialize a TCP server.
// SendMessage must succesfully deliver a message.
func TestTCPServer(t *testing.T) {
	// create TCPServer and add a message handler
	tcp := new(TCPServer)
	tmh := new(TestMsgHandler)
	tcp.MessageHandlers = make(map[byte]common.MessageHandler)
	tcp.MessageHandlers[1] = tmh
	err := tcp.InitServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}

	// send a message to be echoed
	m := common.Message{common.Address{0, "localhost", 9988}, []byte("\x01hello, world!")}
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
	m = common.Message{common.Address{0, "localhost", 9988}, []byte("\xFFwon't be seen")}
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
