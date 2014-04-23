package client

import (
	"common"
	"sync"
	"quorum"
)

type SegmentHandler struct {
	lock sync.Mutex
	file quorum.File
	Id common.Identifier
	segmap map[uint8][]byte
	done chan bool // Value doesn't matter, but could be useful
}

func NewSegmentHandler(file quorum.File, messageSender common.MessageSender) *SegmentHandler {
	c := new(SegmentHandler)
	c.Id = 128 // Arbitrary value
	c.segmap = make(map[uint8][]byte)
	c.sender = messageSender
	c.done = make(chan bool)
	c.SendRequests(messageSender)
	return c
}

func (sh SegmentHandler) SendRequests(sender common.MessageSender) {
	for index, participants := range sh.file.Chunkdistribution {
		for _, participant := range participants {
			mesg := NewSegmentRequest(participant.Address, index)
			sender.SendMessage(mesg)
		}
	}
}

func (sh SegmentHandler) Identifier() common.Identifier {
	return sh.Id
}

func (sh *SegmentHandler) valid(index uint8, segment []byte) bool {
	// Simple check 
	if len(file.Chunkdistribution) < index {
		return false
	}
	return true
}

func (sh *SegmentHandler) HandleMessage(mesg []byte) {
	if len(mesg) <= 1 {
		return
	}
	index := uint8(mesg[0])
	segment := mesg[1:]
	if sh.valid(index, segment) {
		sh.lock.Lock()
		defer sh.lock.Unlock()
		if _, in := sh.segmap[index]; !in {
			sh.segmap[index] = segment
			if len(sh.segmap) == segcount {
				sh.done <- true
			}
		}
	}
}

func (sh *SegmentHandler) SegmentMap() map[uint8][]byte {
	<-sh.done
	sh.lock.Lock()
	defer sh.lock.Unlock()
	return sh.segmap
}
