package common

type ZeroNetwork struct {
	messages []*Message
}

func (z *ZeroNetwork) Address() (a Address) {
	return
}

func (z *ZeroNetwork) RegisterHandler(handler interface{}) (i Identifier) {
	return
}

func (z *ZeroNetwork) SendMessage(m *Message) error {
	z.messages = append(z.messages, m)
	return nil
}

func (z *ZeroNetwork) RecentMessage(i int) *Message {
	if i < len(z.messages) {
		return z.messages[i]
	}
	return nil
}

func NewZeroNetwork() *ZeroNetwork {
	return new(ZeroNetwork)
}
