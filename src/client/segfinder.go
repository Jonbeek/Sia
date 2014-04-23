package client

import (
	"common"
)

type SegFinder struct {
	filehash string
	Id common.Identifier
	// seginfo something
	done chan bool
}

func NewSegFinder(hash string, addr common.Address, sender common.MessageSender) *SegFinder {
	sf := new(SegFinder)
	sf.filehash = hash
	sf.Id = 127 // Arbitrary
	go sf.FileInfo(addr, sender)
	return sf
}

func (sf *SegFinder) FileInfo(addr common.Address, sender common.MessageSender) {
	// Send the request
	// Don't do much else
}

func (sf SegFinder) Identifier() {
	return sf.Id
}

func (sf *SegFinder) HandleMessage(mesg []byte) {
	// Coerce into the segment location type
	done <- true
}

func (sf *SegFinder) SegmentLocation() {
	<-done
	// return seginfo
