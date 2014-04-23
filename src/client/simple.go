package client

import (
	"common"
	"os"
)

// Thank you Luke
type SimpleClient struct {
	Id Identifier
	file *os.File
	done chan bool
	recv chan bool
}

func NewSimpleClient(file *os.File) {
	sc := new(SimpleClient)
	sc.file = file
	sc.done = make(chan bool, 1)
	sc.recv = make(chan bool, 1)
	sc.recv <- true
}

func (sc SimpleClient) Identifier() common.Identifier {
	return sc.Id
}

func (sc *SimpleClient) HandleMessage(payload []byte) {
	// Only do it once.
	<-sc.recv
	sc.file.Write(payload)
	sc.done <- true
}

func (sc *SimpleClient) Wait() {
	<-sc.done
}

