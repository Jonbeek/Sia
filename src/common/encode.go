package common

func TransactionNetworkMessage(t Transaction) NetworkMessage {
	return NetworkMessage{t.SwarmId(), t.TransactionId(), "", t.MarshalString()}
}

func BlockNetworkMessage(b Block) NetworkMessage {
	return NetworkMessage{b.SwarmId(), "", b.BlockId(), b.MarshalString()}
}
