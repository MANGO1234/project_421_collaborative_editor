// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

package network

import (
	"../util"
	"github.com/satori/go.uuid"
	"net"
	"bufio"
	"fmt"
	"sync"
	"encoding/json"
)

// data structures

type Node struct {
	Addr    string
	Quitted bool // node is forever quitted if it voluntarily leave the network

	// fields below are for local communication purpose
	connected bool         // true if current node is connected to target node
	in        *ConnWrapper
	out       *ConnWrapper 
}

type ConnWrapper struct {
	ConnReader *util.MessageReader
	ConnWriter *util.MessageWriter
}

type NodeMetaData struct {
	Id   string
	Addr string
}

type NodeMap struct {
	sync.RWMutex
	Nodes map[string]Node
}

// global variables
var nodeMetaData NodeMetaData                       // keeps track of current node's information
var nodeMap = NodeMap{Nodes: make(map[string]Node)} // keeps track of all nodes in the system

// message type
var RegMsg string = "registration"



// listen for incoming connection
func listenForConn(listener *net.TCPListener) {
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

	if err == nil {
		newUUID := uuid.NewV1().String()
		nodeMetaData = NodeMetaData{newUUID, lAddr.String()}
		go listenForConn(listener)
	}

	return err
}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	conn, err := net.Dial("tcp", remoteAddr)
	msgWriter := util.MessageWriter{bufio.NewWriter(conn)}

	msg, _ := json.Marshal(nodeMetaData)
	msgWriter.WriteMessage2(RegMsg, msg)

	

	return err
}

func Broadcast() {

}
