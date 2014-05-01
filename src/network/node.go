package network

import (
	"common"
	"encoding/binary"
	"net"
	"strconv"
)

// TCPServer is a MessageSender that communicates over TCP.
type TCPServer struct {
	Addr            common.Address
	MessageHandlers []common.MessageHandler
	Listener        net.Listener
}

func (tcp *TCPServer) Address() common.Address {
	return tcp.Addr
}

// AddMessageHandler adds a MessageHandler to the MessageHandlers slice.
// It creates an identifier associated with that MessageHandler, and returns
// an Address incorporating the identifier.
func (tcp *TCPServer) AddMessageHandler(mh common.MessageHandler) common.Address {
	tcp.MessageHandlers = append(tcp.MessageHandlers, mh)
	addr := tcp.Addr
	addr.Id = common.Identifier(len(tcp.MessageHandlers) - 1)
	return addr
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
	// bytes 0:3 are the payload length
	// byte 4 is the destination identifier
	// the remainder is the payload
	payloadLength := make([]byte, 4)
	binary.PutUvarint(payloadLength, uint64(len(m.Payload)))
	stream := append(payloadLength, byte(m.Destination.Id))
	stream = append(stream, m.Payload...)

	// transmit stream
	_, err = conn.Write(stream)
	if err != nil {
		return
	}

	return
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
	// MessageHandlers[0] is reserved for the MessageHandler of the TCPServer
	tcp.MessageHandlers = make([]common.MessageHandler, 1)
	tcp.Listener = tcpServ

	go tcp.serverHandler()
	return
}

// Close closes the connection associated with the TCP server.
// This causes tcpServ.Accept() to return an err, ending the serverHandler process
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
			conn.Close()
		}
	}
}

// clientHandler reads data sent by a client and processes it.
func (tcp *TCPServer) clientHandler(conn net.Conn) {
	// read first 1024 bytes
	buffer := make([]byte, 1024)

	// read first 1024 bytes
	b, err := conn.Read(buffer)
	if err != nil {
		return
	}

	// split message into payload length, identifier, and payload
	payloadLength, _ := binary.Uvarint(buffer[:4])
	id := int(buffer[4])
	payload := make([]byte, b-5)
	copy(payload, buffer[5:b])

	// read rest of payload, 1024 bytes at a time
	// TODO: add a timeout
	bytesRead := len(payload)
	for uint64(bytesRead) != payloadLength {
		b, err = conn.Read(buffer)
		if err != nil {
			return
		}
		payload = append(payload, buffer[:b]...)
		bytesRead += b
	}

	// Message sent directly to TCPServer
	// for now, just send it to the first message handler
	if id == 0 {
		tcp.MessageHandlers[1].HandleMessage(payload)
		return
	}

	// look up message handler and call it
	if id < len(tcp.MessageHandlers) {
		go tcp.MessageHandlers[id].HandleMessage(payload)
	}
}
