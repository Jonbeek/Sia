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

	return MarshalTransaction("NodeAlive", string(b))
}

type transactionwrap struct {
	Type  string
	Value string
}

func MarshalTransaction(Type string, Value string) string {
	b, err := json.Marshal(transactionwrap{Type, Value})
	if err != nil {
		panic("Unable to marshal transactionwrap, should be impossible " + err.Error())
	}
	return string(b)
}

func UnmarshalTransaction(b string) (common.Transaction, error) {
	t := new(transactionwrap)
	err := json.Unmarshal([]byte(b), t)
	switch t.Type {
	case "NodeAlive":
		c := new(NodeAlive)
		err = json.Unmarshal([]byte(t.Value), c)
		if err != nil {
			return nil, err
		}
		return c, nil
	case "Heartbeat":
		h := new(HeartbeatTransaction)
		err = json.Unmarshal([]byte(t.Value), h)
		if err != nil {
			return nil, err
		}
		return h, nil
	default:
		return nil, errors.New("Unknown transaction type")
	}
}
