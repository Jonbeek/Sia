package swarm

import (
	"encoding/json"
}

type HeartbeatTransaction struct {
	Id string
	Swarm string
	Stage1 string
	Stage2 string
}

func NewHeartbeat(Swarm, Stage1, Stage2 string) (h *HeartbeatTransaction) {
	h = new(HeartbeatTransaction)
	h.Swarm = Swarm
	h.Stage1 = Stage1
	h.Stage2 = Stage2
	Id, err := EntropyBytes()
	if err != nil {
		panic("Failed to make new heartbeat transaction" + err.Error())
	}
	h.Id = Id
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

	return MarshalTransaction("Heartbeat", string(b))
}

