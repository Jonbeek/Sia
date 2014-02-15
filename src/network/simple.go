package network

type SimpleMultiplexes struct {
    hosts []chan NetworkObject
}

func (s *SimpleMultiplexer) AddListener(SwarmId string, c chan NetworkObject) {
    s.hosts = append(s.hosts)
}

func (s *SimpleMultiplexer) SendNetworkObject(o NetworkObject) {
    for _, c := range s.hosts {
        go func() {
            c <- o
        }()
    }
}

func (s *SimpleMultipexer) Listen(addr string) {
    panic("Not implemented")
}

func (s *SimpleMultiplexer) Connect(addr string) {
    panic("Not implemented")
}
