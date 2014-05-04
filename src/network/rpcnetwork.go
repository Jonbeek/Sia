package network

import (
	"common"
	"net"
	"net/rpc"
	"strconv"
)

// RPCServer is a MessageRouter that communicates using RPC over TCP.
type RPCServer struct {
	addr     common.Address
	rpcServ  *rpc.Server
	listener net.Listener
}

func (rpcs *RPCServer) Address() common.Address {
	return rpcs.addr
}

// RegisterHandler registers a message handler to the RPC server.
func (rpcs *RPCServer) RegisterHandler(handler interface{}) {
	rpcs.rpcServ.Register(handler)
}

// NewRPCServer creates and initializes a server that listens for TCP connections on a specified port.
// It then spawns a serverHandler with a specified message.
// It is the serverHandler's responsibility to close the TCP connection.
func NewRPCServer(port int) (rpcs *RPCServer, err error) {
	tcpServ, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return
	}

	rpcs = &RPCServer{
		common.Address{0, "localhost", port},
		rpc.NewServer(),
		tcpServ,
	}

	go rpcs.serverHandler()
	return
}

// Close closes the connection associated with the TCP server.
// This causes tcpServ.Accept() to return an err, ending the serverHandler process
func (rpcs *RPCServer) Close() {
	rpcs.listener.Close()
}

// serverHandler accepts incoming RPCs, serves them, and closes the connection.
func (rpcs *RPCServer) serverHandler() {
	for {
		conn, err := rpcs.listener.Accept()
		if err != nil {
			return
		} else {
			go func() {
				rpcs.rpcServ.ServeConn(conn)
				conn.Close()
			}()
		}
	}
}

// SendRPCMessage (synchronously) delivers an RPCMessage to its recipient and returns any errors.
func SendRPCMessage(m *common.RPCMessage) error {
	conn, err := rpc.Dial("tcp", net.JoinHostPort(m.Destination.Host, strconv.Itoa(m.Destination.Port)))
	if err != nil {
		return err
	}
	return conn.Call(m.Proc, m.Args, m.Reply)
}

// SendAsyncRPCMessage (asynchronously) delivers an RPCMessage to its recipient.
// It returns a *Call, which contains the fields "Done channel" and "Error error".
func SendAsyncRPCMessage(m *common.RPCMessage) *rpc.Call {
	conn, err := rpc.Dial("tcp", net.JoinHostPort(m.Destination.Host, strconv.Itoa(m.Destination.Port)))
	d := make(chan *rpc.Call, 1)
	if err != nil {
		// make a dummy *Call
		errCall := &rpc.Call{"", nil, nil, err, d}
		errCall.Done <- nil
		return errCall
	}
	return conn.Go(m.Proc, m.Args, m.Reply, d)
}
