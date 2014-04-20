package network

import (
	"bytes"
	"common"
	"net"
	"strconv"
)

// TCPServer is a MessageSender that communicates over TCP.
// MessageHandlers is a map of Identifiers to MessageHandler interfaces.
type TCPServer struct {
	Addr            common.Address
	MessageHandlers map[common.Identifier]common.MessageHandler
	Listener        net.Listener
}

// Address returns the address of the server
func (tcp *TCPServer) Address() common.Address {
	return tcp.Addr
}

// SendMessage transmits the payload of a message to its intended recipient.
// It marshalls the Message struct using a length-prefix scheme.
// It does not wait for a response.
func (tcp *TCPServer) SendMessage(m *common.Message) (err error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(m.Destination.Host, strconv.Itoa(m.Destination.Port)))
	if err != nil {
		return
	}
	defer conn.Close()

	// construct stream to be transmitted
	// 0x00 is used as a delimiter
	stream := bytes.Join([][]byte{
		[]byte(strconv.Itoa(len(m.Payload))), // length of payload
		[]byte{byte(m.Destination.Id)},       // identifier
		m.Payload,                            // payload
	}, []byte{0xFF})

	// transmit stream
	_, err = conn.Write(stream)
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
	tcp.Listener = tcpServ

	go tcp.serverHandler()
	return
}

// Close closes the connection associated with the TCP server.
// This causes tcpServ.Accept to return an err, ending the serverHandler process
func (tcp *TCPServer) Close() {
	tcp.Listener.Close()
}

// serverHandler accepts incoming connections and spawns a clientHandler for each.
func (tcp *TCPServer) serverHandler() {
	for {
		conn, err := tcp.Listener.Accept()
		if err != nil {
			return
		} else {
			tcp.clientHandler(conn)
		}
	}
}

// clientHandler reads data sent by a client and processes it.
func (tcp *TCPServer) clientHandler(conn net.Conn) {
	var payload []byte
	buffer := make([]byte, 1024)

	// read first 1024 bytes
	b, err := conn.Read(buffer)
	if err != nil {
		return
	}

	// split message into payload length, identifier, and payload
	splitMessage := bytes.SplitN(buffer[:b], []byte{0xFF}, 3)
	payloadLength, _ := strconv.Atoi(string(splitMessage[0]))
	id := common.Identifier(splitMessage[1][0])
	payload = splitMessage[2]

	// read rest of payload, 1024 bytes at a time
	// TODO: add a timeout
	bytesRead := len(payload)
	for bytesRead != payloadLength {
		b, err = conn.Read(buffer)
		if err != nil {
			return
		}
		payload = append(payload, buffer[:b]...)
		bytesRead += b
	}

	// look up message handler and call it
	handler, exists := tcp.MessageHandlers[id]
	if exists {
		handler.HandleMessage(payload)
	}
	// todo: decide on behavior when encountering uninitialized Identifier
}
