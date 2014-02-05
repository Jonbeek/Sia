type Transaction interface {
    SwarmId() string
    TransactionId() string
    MarshalString() string
}

type Block interface {
    SwarmId() string
    BlockId() string
    MarshalString() string
}

type NetworkObject struct {
    SwarmId string
    TransactionId  string
    BlockId string
    payload string
}

type NetworkMultiPlexer interface {
    AddListener(Swarmid string, chan NetworkObject)
    SendTransaction(t Transaction)
    SendBlock(b Block)
    Listen(addr string)
    Connect(addr string)
}
