package swarm

import (
	"common"
	"encoding/json"
)

type HeartbeatList struct {
	Id         string
	Blockchain string
	Host       string

	Signatures map[string]string
	Heartbeats map[string]*Heartbeat
}

func NewHeartbeatList(blockchain, host string,
	heartbeats map[string]*Heartbeat,
	signatures map[string]string) *HeartbeatList {
	h := new(HeartbeatList)
	h.Id, _ = common.RandomString(8)
	h.Host = host
	h.Blockchain = blockchain
	h.Signatures = signatures
	h.Heartbeats = heartbeats
	return h
}

func (h *HeartbeatList) SwarmId() string {
	return h.Blockchain
}

func (h *HeartbeatList) UpdateId() string {
	return h.Id
}

func (h *HeartbeatList) Type() string {
	return "HeartbeatList"
}

func (h *HeartbeatList) MarshalString() string {
	w, err := json.Marshal(h)
	if err != nil {
		panic("Unable to marshal HeartBeatTransaction, this should not happen" + err.Error())
	}

	return string(w)
}
