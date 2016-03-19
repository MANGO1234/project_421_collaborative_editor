// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

package network

import (
	"../util"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"net"
	"sync"
)

// data structures

type Node struct {
	Addr    string
	Quitted bool // node is forever quitted if it voluntarily leave the network

	// fields below are for local communication purpose
	connected bool // true if current node is connected to target node
	in        *ConnWrapper
	out       *ConnWrapper
}

type ConnWrapper struct {
	ConnReader *util.MessageReader
	ConnWriter *util.MessageWriter
}

type NodeMetaData struct {
	Id      string
	Addr    string
	NodeMap map[string]Node
}

// global variables
var nodeMetaData = NodeMetaData{NodeMap: make(map[string]Node)} // keeps track of nodeMetaData
var mapLock *sync.RWMutex

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
	msgWriter := util.MessageWriter{bufio.NewWriter(conn)}
	msgReader := util.MessageReader{bufio.NewReader(conn)}
	msgInType, m, err := msgReader.ReadMessage2()
	handleError(err)

	var msgIn NodeMetaData
	err = json.Unmarshal(m, &msgIn)

	fmt.Println("received message Type: ", msgInType) //
	//add this node to nodeMap

	_, ok := nodeMetaData.NodeMap[msgIn.Id]
	if !ok {
		nodeMetaData.NodeMap[msgIn.Id] = Node{Addr: msgIn.Addr, Quitted: false, connected: true}
	}

	// reply
	if msgInType == RegMsg {
		msg, _ := json.Marshal(nodeMetaData)
		msgWriter.WriteMessage2(RegMsg, msg)
	}
	// TODO: connect back to node

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
	fmt.Println("listening on ", lAddr.String())

	if err == nil {
		newUUID := uuid.NewV1().String()
		nodeMetaData = NodeMetaData{newUUID, lAddr.String(), make(map[string]Node)}
		go listenForConn(listener)
	}

	return err
}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}

	msgWriter := util.MessageWriter{bufio.NewWriter(conn)}
	msgReader := util.MessageReader{bufio.NewReader(conn)}

	// send registration information
	msg, _ := json.Marshal(nodeMetaData)

	err = msgWriter.WriteMessage2(RegMsg, msg)
	handleError(err)

	// receive registration information
	msgType, msgBuff, err := msgReader.ReadMessage2()
	handleError(err)
	var newNodeData NodeMetaData
	if msgType == RegMsg {
		json.Unmarshal(msgBuff, &newNodeData)
	}
	fmt.Println("received", msgType)

	// save new node to map
	var newNode Node
	newNode.Addr = newNodeData.Addr
	newNode.connected = true
	outConnWrapper := ConnWrapper{&msgReader, &msgWriter}
	newNode.out = &outConnWrapper
	
	addNodeToMap(newNode, newNodeData.Id)
	
	handleNewNodes(newNodeData.NodeMap)

	return err
}

func Broadcast() {

}

func addNodeToMap(nodeData Node, nodeId string) {
	nodeMetaData.NodeMap[nodeId] = nodeData
}

func handleNewNodes(receivedNodeMap map[string]Node) {

	for key, value := range receivedNodeMap {
		_, ok := nodeMetaData.NodeMap[key]
		if !ok {
			addNodeToMap(value, key)
			// TODO: Connect to the added node.
		}
	}
}

func handleError(error error) {

}
