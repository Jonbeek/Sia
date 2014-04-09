package network

import (
	"net"
	"strconv"
)

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

/* process incoming messages */
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
	return
}

/* send a message to another node and return the response */
func SendMessage(host string, port int, message []byte) (resp []byte, err error) {
	/* open connection */
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
    b, err := conn.Read(buffer)
    resp = buffer[:b]
	return
}

/* create node and begin listening on specified port */
func InitNode(port int, data []byte) (err error) {
	tcpServ, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	go serverHandler(tcpServ, data)
	return
}
