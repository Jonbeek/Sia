package common

func TransactionNetworkObject(t Transaction) NetworkObject {
	return NetworkObject{t.SwarmId(), t.TransactionId(), "", t.MarshalString()}
}
func BlockNetworkObject(b Block) NetworkObject {
	return NetworkObject{b.SwarmId(), "", b.BlockId(), b.MarshalString()}
}
