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
}

func NewSimpleClient(file *os.File) {
	sc := new(SimpleClient)
	sc.file = file
	sc.done = make(chan bool, 1)

func (sc SimpleClient) Identifier() common.Identifier {
	return sc.Id
}

func (sc *SimpleClient) HandleMessage(payload []byte) {
	sc.file.Write(payload)
	sc.done <- true
}

func (sc *SimpleClient) Wait() {
	<-sc.done
}

