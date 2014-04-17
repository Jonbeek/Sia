package common

import (
	"testing"
)

// use all functions of ZeroNetwork, expecting no nil returns and no errors
func TestZeroNetwork(t *testing.T) {
	zeroNet := NewZeroNetwork()
	if zeroNet == nil {
		t.Fatal("NewZeroNetwork cannot return nil")
	}

	m := new(Message)
	m.Destination = zeroNet.Address()
	m.Payload = make([]byte, 1)
	m.Payload[0] = 3

	err := zeroNet.SendMessage(m)
	if err != nil {
		t.Fatal("ZeroNetwork.SendMessage cannot return and error")
	}
}
