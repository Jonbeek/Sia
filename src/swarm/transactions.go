package swarm

import (
	"common"
	"encoding/json"
	"errors"
)

type NodeAlive struct {
	Node      string
	Signature string
	Swarm     string
	Id        string
}

func NewNodeAlive(Node string, Swarm string) (n *NodeAlive) {
	n = new(NodeAlive)
	n.Node = Node
	n.Swarm = Swarm
	b, _ := EntropyGeneration() //Should be the hash of the message, but eh
	n.Id = string(b)
	n.Signature = "TODO"
	return
}

func (n *NodeAlive) SwarmId() string {
	return n.Swarm
}

func (n *NodeAlive) UpdateId() string {
	return n.Id
}

func (n *NodeAlive) Type() string {
	return "NodeAlive"
}

func (n *NodeAlive) MarshalString() string {
	b, err := json.Marshal(n)
	if err != nil {
		panic("Unable to marshal NodeAlive, should be impossible " + err.Error())
	}

	return string(b)
}

type transactionwrap struct {
	Type  string
	Value string
}

func UnmarshalUpdate(m common.NetworkMessage) (common.Update, error) {
	var u common.Update

	switch m.Type {
	case "NodeAlive":
		u = new(NodeAlive)
	case "Heartbeat":
		u = new(Heartbeat)
	case "Block":
		u = new(Block)
	default:
		return nil, errors.New("Unknown transaction type")
	}

	err := json.Unmarshal([]byte(m.Payload), u)
	return u, err

}
