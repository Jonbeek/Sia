package swarm

import (
	"common"
	"encoding/json"
	"errors"
)

func UnmarshalUpdate(m common.NetworkMessage) (common.Update, error) {
	var u common.Update

	switch m.Type {
	case "Heartbeat":
		u = new(Heartbeat)
	case "HeartbeatList":
		u = new(HeartbeatList)
	case "Block":
		u = new(Block)
	default:
		return nil, errors.New("Unknown transaction type")
	}

	err := json.Unmarshal([]byte(m.Payload), u)
	return u, err

}
