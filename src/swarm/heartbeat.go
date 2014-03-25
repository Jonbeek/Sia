package swarm

import (
	"common"
	"encoding/json"
	"swarm/record"
)

type Heartbeat struct {
	Id string

	Blockchain string
	Host       string

	EntropyStage1   string
	EntropyStage2   string
	FileProofStage1 string
	FileProofStage2 string

	Recordsenc []string
}

func NewHeartbeat(swarm, Host, Stage1, Stage2 string, records []common.Record) (h *Heartbeat) {
	h = new(Heartbeat)
	h.Blockchain = swarm
	h.Host = Host
	h.EntropyStage1 = Stage1
	h.EntropyStage2 = Stage2
	h.Id, _ = common.RandomString(8)

	h.Recordsenc = nil
	for _, r := range records {
		h.Recordsenc = append(h.Recordsenc, record.Encode(r))
	}

	return
}

func (h *Heartbeat) SwarmId() string {
	return h.Blockchain
}

func (h *Heartbeat) UpdateId() string {
	return h.Id
}

func (h *Heartbeat) Type() string {
	return "Heartbeat"
}

func (h *Heartbeat) MarshalString() string {
	w, err := json.Marshal(h)
	if err != nil {
		panic("Unable to marshal HeartbeatTransaction, this should not happen" + err.Error())
	}

	return string(w)
}

func (h *Heartbeat) GetRecords() []common.Record {
	r := make([]common.Record, 0, len(h.Recordsenc))

	for _, rd := range h.Recordsenc {
		r = append(r, record.Decode(rd))
	}

	return r
}
