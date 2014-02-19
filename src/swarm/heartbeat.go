package swarm

import (
	"common"
	"common/crypto"
	"encoding/json"
	"errors"
)

type Heartbeat struct {
	Id          string
	Blockchain  string
	ParentBlock string

	EntropyStage1   string
	EntropyStage2   string
	FileProofStage1 string
	FileProofStage2 string

	Transactions map[string]Update
}

func (s *StateSteady) NewHeartbeat() (hb Heartbeat, err error) {
	// temporary, eventually heartbeats will not have an id
	// instead they will be id'd by their hash
	hb.Id, _ = common.RandomString(8)

	// create entropy for the block
	newEntropy, err := EntropyGeneration()
	if err != nil {
		return
	}

	// put the hash of the new entropy as EntropyStage1
	hb.EntropyStage1 = string(crypto.Hash(newEntropy))

	// put stored entropy string from previous heartbeat as hb.EntropyStage2
	hb.EntropyStage2 = s.SecretEntropy

	// save entropy string for hb.EntropyStage2 of next heartbeat
	s.SecretEntropy = string(newEntropy)

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
func (s *StateSteady) AddHeartbeat(hb Heartbeat) (err error) {
	err = hb.verify()
	if err == nil {
		s.ActiveBlock.Heartbeats[hb.Id] = hb
	}

	// if no error, send an ack

	return
}

func (s *StateSteady) RemoveHeartbeat(id string) {
	delete(s.ActiveBlock.Heartbeats, id)
}

func (hb *Heartbeat) verify() (err error) {
	// check signature, make sure all objects seem reasonable
	err = errors.New("Verification not implemented")
	return
}

func (hb *Heartbeat) SwarmId() string {
	return hb.Blockchain
}

func (hb *Heartbeat) TransactionId() string {
	return hb.Id
}

func (h *Heartbeat) MarshalString() string {
	w, err := json.Marshal(h)
	if err != nil {
		panic("Unable to marshal Heartbeat, this should not happen" + err.Error())
	}

	return MarshalTransaction("HeartBeat", string(w))
}

func VerifyHeartbeat(prevBlock *Block, h *Heartbeat) {
	// Just return true for now
	// DANGEROUS
	return
}
