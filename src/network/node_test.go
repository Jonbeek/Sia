package network

import (
    "testing"
)

func TestNetworkNode(t *testing.T) {
    err := InitNode(9988, []byte("foo"))
    if err != nil {
        t.Fatal("Failed to initialize node:", err)
    }
    resp, err := SendMessage("localhost", 9988, []byte("req"))
    if err != nil {
        t.Fatal("Failed to send message:", err)
    }
    if string(resp) != "foo" {
        t.Fatal("Bad response: expected 'foo', got", resp)
    }
    resp, err = SendMessage("localhost", 9988, []byte("cmd"))
    if string(resp) != "unrecognized command \"cmd\"" {
        t.Fatal("Bad response, expected 'unrecognized command \"cmd\"', got", resp)
    }
}
