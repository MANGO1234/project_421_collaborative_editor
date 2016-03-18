// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up: 
// $ go get github.com/satori/go.uuid

package network

import(
	"../util"
	"net"
	"github.com/satori/go.uuid"
)

// data structures

type Node struct {
	Addr string
	Quitted bool  // node is forever quitted if it voluntarily leave the network

	// fields below are for local communication purpose
	connected bool // true if current node is connected to target node
	in   *net.TCPConn // tcpConn for listening from that node
	out  *net.TCPConn // tcpConn for writing to that node
}

type NodeMetaData = struct {
	Id    string
	Addr  string
}

type NodeMap = struct {
	sync.RWMutex
	Nodes map[string]Node
}

// global variables
var nodeMetaData NodeMetaData // keeps track of current node's information
var nodeMap = NodeMap{Nodes: make(map[string]Node)} // keeps track of all nodes in the system


// listen for incoming connection
func listenForConn(listener *TCPListener) {
	for {
		conn, _ := listener.Accept()
		handleConn(conn)
	}
}

// handle incoming connection
func handleConn(conn net.Conn) {
	// TODO: read incoming messages. 
	// Deal with following cases:


	// new node joining : 
	// send saved nodeMap to that node, add that node to nodeMap, merge treedoc if necessary.

	// known node quitting :
	// set map[node].Quitted = true.


	// editing treedoc:
	// modify treedoc


	// ** for any received messages, broadcast may be required depending on the situation
}

// initialize local network listener 
func Initialize(addr string) error {
	lAddr, err := net.ResolveTCPAddr("tcp", addr)
	listener, err := net.ListenTCP("tcp", lAddr)

	if err == nil{ 
		newUUID := uuid.NewV1().String()
		nodeMataData = NodeMetaData{newUUID, lAddr}
		go listenForConn(listener)
	}

	return err
}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	// TODO
	return nil
}

func Broadcast() {

}