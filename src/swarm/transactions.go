package swarm

import (
	"encoding/json"
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
	b, _ := EntropyBytes() //Should be the hash of the message, but eh
	n.Id = string(b)
	n.Signature = "TODO"
	return
}

func (n *NodeAlive) SwarmId() string {
	return n.Swarm
}

func (n *NodeAlive) TransactionId() string {
	return n.Id
}

func (n *NodeAlive) MarshalString() string {
	b, err := json.Marshal(n)
	if err != nil {
		panic("Unable to marshal NodeAlive, should be impossible " + err.Error())
	}
	return string(b)
}
