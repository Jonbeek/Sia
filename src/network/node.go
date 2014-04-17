package network

import (
	"common"
	"net"
	"strconv"
)

// TCPServer is a MessageSender that communicates over TCP.
// MessageHandlers is a map of bytecodes to MessageHandler interfaces.
type TCPServer struct {
	Addr            common.Address
	MessageHandlers map[byte]common.MessageHandler
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
	_, err = conn.Write(m.Payload)
	if err != nil {
		return
	}
	return
}

// InitServer initializes a server that listens for TCP connections on a specified port.
// It then spawns a serverHandler with a specified message.
// It is the serverHandler's responsibility to close the TCP connection.
func (tcp *TCPServer) InitServer(port int) (err error) {
	tcpServ, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}

	tcp.Addr = common.Address{0, "localhost", port}

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
	handler, exists := tcp.MessageHandlers[buffer[0]]
	if exists {
		handler.HandleMessage(buffer[1:b])
	}
}
