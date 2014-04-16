package network

import (
	"net"
	"strconv"
)

// serverHandler accepts incoming connections and spawns a clientHandler for each.
func serverHandler(tcpServ net.Listener, data []byte) {
	defer tcpServ.Close()

	for {
		conn, err := tcpServ.Accept()
		if err != nil {
			return
		} else {
			go clientHandler(conn, data)
		}
	}
}

// clientHandler reads data sent by a client and processes it.
// It then sends a response based on the client's message.
func clientHandler(conn net.Conn, data []byte) {
	buffer := make([]byte, 1024)
	b, err := conn.Read(buffer)
	if err != nil {
		return
	}

	cmd := string(buffer[:b])
	switch cmd {
	case "req":
		conn.Write(data)
	default:
		conn.Write([]byte("unrecognized command \"" + cmd + "\""))
	}
}

// SendMessage sends a message over TCP to a specified host and port.
// It then waits for a response, and returns it after closing the connection.
func SendMessage(host string, port int, message []byte) (resp []byte, err error) {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = conn.Write(message)
	if err != nil {
		return
	}

	buffer := make([]byte, 1024)
	// TODO: add a timeout here
	b, err := conn.Read(buffer)
	resp = buffer[:b]
	return
}

// InitNode initializes a server that listens for TCP connections on a specified port.
// It then spawns a serverHandler with a specified message.
// It is the serverHandler's responsibility to close the TCP connection.
func InitNode(port int, data []byte) (err error) {
	tcpServ, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}

	go serverHandler(tcpServ, data)
	return
}
