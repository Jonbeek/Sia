package swarm

import (
	"common"
	"encoding/json"
)

type HeartBeatTransaction struct {
	Id        string
	Swarm     string
	Host      string
	Stage1    string
	Stage2    string
	Prevblock string
}

func NewHeartBeat(prevState *Block, Host, Stage1, Stage2 string) (h *HeartBeatTransaction) {
	h = new(HeartBeatTransaction)
	h.Swarm = prevState.SwarmId()
	h.Stage1 = Stage1
	h.Stage2 = Stage2
	h.Id, _ = common.RandomString(8)
	h.Prevblock = prevState.Id
	return
}

func (h *HeartBeatTransaction) SwarmId() string {
	return h.Swarm
}

func (h *HeartBeatTransaction) TransactionId() string {
	return h.Id
}

func (h *HeartBeatTransaction) MarshalString() string {
	w, err := json.Marshal(h)
	if err != nil {
		panic("Unable to marshal HeartBeatTransaction, this should not happen" + err.Error())
	}

	return MarshalTransaction("HeartBeat", string(w))
}

func (h *HeartBeatTransaction) GetStage2() string {
	return h.Stage2
}

func VerifyHeartBeat(prevBlock *Block, h *HeartBeatTransaction) {
	// Just return true for now
	// DANGEROUS
	return
}
