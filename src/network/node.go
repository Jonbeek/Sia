package network

import (
	"common"
	"net"
	"strconv"
)

// TCPServer is a MessageSender that communicates over TCP.
// MessageHandlers is a map of Identifiers to MessageHandler interfaces.
type TCPServer struct {
	Addr            common.Address
	MessageHandlers map[common.Identifier]common.MessageHandler
}

// Address returns the address of the server
func (tcp *TCPServer) Address() common.Address {
	return tcp.Addr
}

// SendMessage transmits the payload of a message to its intended recipient.
// It returns without waiting for a response.
func (tcp *TCPServer) SendMessage(m *common.Message) (err error) {
	conn, err := net.Dial("tcp", m.Destination.Host+":"+strconv.Itoa(m.Destination.Port))
	if err != nil {
		return
	}
	defer conn.Close()

	// append identifier to front of payload
	payload := append([]byte{byte(m.Destination.Id)}, m.Payload...)
	_, err = conn.Write(payload)
	if err != nil {
		return
	}
	return
}

// AddMessageHandler adds a MessageHandler to the MessageHandlers map
// If the key already has a MessageHandler associated with it, it is overwritten.
func (tcp *TCPServer) AddMessageHandler(mh common.MessageHandler) {
	tcp.MessageHandlers[mh.Identifier()] = mh
}

// NewTCPServer creates and initializes a server that listens for TCP connections on a specified port.
// It then spawns a serverHandler with a specified message.
// It is the serverHandler's responsibility to close the TCP connection.
func NewTCPServer(port int) (tcp *TCPServer, err error) {
	tcp = new(TCPServer)
	tcpServ, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}

	// initialize struct fields
	tcp.Addr = common.Address{0, "localhost", port}
	tcp.MessageHandlers = make(map[common.Identifier]common.MessageHandler)

	go tcp.serverHandler(tcpServ)
	return
}

// serverHandler accepts incoming connections and spawns a clientHandler for each.
func (tcp *TCPServer) serverHandler(tcpServ net.Listener) {
	defer tcpServ.Close()

	for {
		conn, err := tcpServ.Accept()
		if err != nil {
			return
		} else {
			tcp.clientHandler(conn)
		}
	}
}

// clientHandler reads data sent by a client and processes it.
func (tcp *TCPServer) clientHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	b, err := conn.Read(buffer)
	if err != nil {
		return
	}
	// look up message handler and call it
	// eventually this will use an unmarshalling function
	handler, exists := tcp.MessageHandlers[common.Identifier(buffer[0])]
	if exists {
		handler.HandleMessage(buffer[1:b])
	}
	// todo: decide on behavior when encountering uninitialized Identifier
}
