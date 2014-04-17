package common

type ZeroNetwork struct {
}

func (z *ZeroNetwork) Address() Address {
	var a Address
	return a
}

func (z *ZeroNetwork) SendMessage(m *Message) error {
	return nil
}

func NewZeroNetwork() *ZeroNetwork {
	return new(ZeroNetwork)
}
