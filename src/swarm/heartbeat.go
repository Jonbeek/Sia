package swarm

import (
	"common/crypto"
	"errors"
	"fmt"
)

type Heartbeat struct {
	Id string

	EntropyStage1   []byte
	EntropyStage2   []byte
	FileProofStage1 []byte
	FileProofStage2 []byte

	Transactions map[string]Update
}

func (s *State) NewHeartbeat() (hb Heartbeat, err error) {
	// create entropy for the block
	newEntropy, err := EntropyGeneration()
	if err != nil {
		return
	}

	// put the hash of the new entropy as EntropyStage1
	hb.EntropyStage1 = crypto.Hash(newEntropy)

	// put stored entropy string from previous heartbeat as hb.EntropyStage2
	hb.EntropyStage2 = make([]byte, EntropyVolume)
	bytesCopied := copy(hb.EntropyStage2, s.SecretEntropy)
	if bytesCopied != EntropyVolume {
		err = fmt.Errorf("Expected to copy %v bytes, copied %v bytes", EntropyVolume, bytesCopied)
		return
	}

	// save entropy string for hb.EntropyStage2 of next heartbeat
	bytesCopied = copy(s.SecretEntropy, newEntropy)
	if bytesCopied != EntropyVolume {
		err = fmt.Errorf("Expected to copy %v bytes, copied %v bytes", EntropyVolume, bytesCopied)
		return
	}

	// figure out which section of the data stack is being selected for storage proof
	// hash that section of data stack, include as FileProofStage1
	// put saved file proof from last block into heartbeat as FileProofStage2
	// put recent file proof into memory for next heartbeat as FileProofStage2

	// include all updates

	// marshall heartbeat
	// sign heartbeat
	// send over network

	// return the heartbeat so you can add it into your own block

	return
}

// this is the type of error that should probably be logged
func (s *State) AddHeartbeat(hb Heartbeat) (err error) {
	err = hb.verify()
	if err == nil {
		s.ActiveBlock.Heartbeats[hb.Id] = hb
	}

	// if no error, send an ack

	return
}

// might change to RH(hb Heartbeat)
func (s *State) RemoveHeartbeat(id string) {
	delete(s.ActiveBlock.Heartbeats, id)
}

func (hb *Heartbeat) verify() (err error) {
	// check signature, make sure all objects seem reasonable
	err = errors.New("Verification not implemented")
	return
}
