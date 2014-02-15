package swarm

import (
	"crypto/sha256"
	"encoding/json"
)

type HeartbeatTransaction struct {
	Id     		string
	Swarm  		string
	Stage1 		string
	Stage2 		string
	Prevblock	*Block
}

func NewHeartbeat(prevState *Block, Swarm, Stage1, Stage2 string) (h *HeartbeatTransaction) {
	h = new(HeartbeatTransaction)
	h.Swarm = prevState.SwarmId()
	h.Stage1 = Stage1
	h.Stage2 = Stage2
	h.Id = EntropyBytes()
	h.Prevblock = prevState
	return
}

func (h *HeartbeatTransaction) SwarmId() string {
	return h.Swarm
}

func (h *HeartbeatTransaction) TransactionId() string {
	return h.Id
}

func (h *HeartbeatTransaction) MarshalString() string {
	w, err := json.Marshal(h)
	if err != nil {
		panic("Unable to marshal HeartbeatTransaction, this should not happen" + err.Error())
	}

	return MarshalTransaction("Heartbeat", string(w))
}

func (h *HeartbeatTransaction) GetStage2() string {
	return h.Stage2
}

func VerifyHeartbeat(prevBlock *Block, h *HeartbeatTransaction) {
	// Just return true for now
	// DANGEROUS
	return true
}

