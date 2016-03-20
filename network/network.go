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
	Connection net.Conn
	ConnReader *util.MessageReader
	ConnWriter *util.MessageWriter
}

type NodeMetaData struct {
	sync.Mutex
	Id      string
	Addr    string
	NodeMap map[string]Node
}

// global variables
var nodeMetaData NodeMetaData = NodeMetaData{NodeMap: make(map[string]Node)} // keeps track of nodeMetaData

// message type
var RegMsg string = "registration"
var LeaveMsg string = "disconnection"

// initialize local network listener
func Initialize(addr string) error {
	lAddr, err := net.ResolveTCPAddr("tcp", addr)
	listener, err := net.ListenTCP("tcp", lAddr)
	fmt.Println("listening on ", lAddr.String())

	if err == nil {
		newUUID := uuid.NewV1().String()
		nodeMetaData.Id = newUUID
		nodeMetaData.Addr = lAddr.String()
		go listenForConn(listener)
	}

	return err
}

// Disconnect from the network voluntarily
func Disconnect() {
	//close all outgoing connection
	for _, value := range nodeMetaData.NodeMap {
		if !value.Quitted {
			// send disconnection information
			err := value.out.ConnWriter.WriteMessage2(LeaveMsg, make([]byte, 100))
			handleError(err)
			// close all out connection
			value.out.Connection.Close()
			fmt.Println("disconnected --- ", value.Addr)
		}
	}
}

// listen for incoming connection to register
func listenForConn(listener *net.TCPListener) {
	for {
		conn, _ := listener.Accept()
		go handleConn(conn)
	}
}

// handle node joining or rejoining
func handleConn(conn net.Conn) {
	// send saved nodeMap to that node, add that node to nodeMap
	msgWriter := util.MessageWriter{bufio.NewWriter(conn)}
	msgReader := util.MessageReader{bufio.NewReader(conn)}
	msgInType, m, err := msgReader.ReadMessage2()
	handleError(err)

	var msgIn NodeMetaData
	err = json.Unmarshal(m, &msgIn)

	fmt.Println("received message Type: ", msgInType) //
	fmt.Println("received node data: ", string(m))

	// reply
	if msgInType == RegMsg {
		msg, _ := json.Marshal(nodeMetaData)
		msgWriter.WriteMessage2(RegMsg, msg)
	}

	//add this node to nodeMap
	nodeMetaData.Lock()
	_, ok := nodeMetaData.NodeMap[msgIn.Id]
	if !ok {
		newNode := Node{Addr: msgIn.Addr, Quitted: false, connected: true}
		addNodeToMap(newNode, msgIn.Id)
		// connect to this node
		connectToHelper(msgIn.Addr)
	}
	handleNewNodes(msgIn.NodeMap)
	nodeMetaData.Unlock()
	for {
		continueRead(msgIn.Id, msgReader)
	}
}

func continueRead(id string, msgReader util.MessageReader) {
	//
	msgInType, _, err := msgReader.ReadMessage2()
	handleError(err)

	if msgInType == LeaveMsg {
		result, ok := nodeMetaData.NodeMap[id]
		if ok && !result.Quitted {

			nodeMetaData.NodeMap[id].out.Connection.Close()
			nodeMetaData.NodeMap[id] = Node{Addr: result.Addr, Quitted: true, connected: false}
			fmt.Println(nodeMetaData.NodeMap[id].Addr, "-----quitted")
		}
	}

}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	nodeMetaData.Lock()
	defer nodeMetaData.Unlock()
	return connectToHelper(remoteAddr)
}

func connectToHelper(remoteAddr string) error {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	fmt.Println("connecting to: ", remoteAddr)
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

	// for checking
	fmt.Println("reply type:", msgType)

	// save new node to map
	var newNode Node
	newNode.Addr = newNodeData.Addr
	newNode.connected = true
	outConnWrapper := ConnWrapper{conn, &msgReader, &msgWriter}
	newNode.out = &outConnWrapper

	addNodeToMap(newNode, newNodeData.Id)

	handleNewNodes(newNodeData.NodeMap)

	return err
}

func GetNetworkMetadata() string {
	nodeMetaData.Lock()
	defer nodeMetaData.Unlock()
	meta, _ := json.Marshal(nodeMetaData)
	return string(meta)
}

func Broadcast() {

}

func addNodeToMap(nodeData Node, nodeId string) {
	nodeMetaData.NodeMap[nodeId] = nodeData
}

func handleNewNodes(receivedNodeMap map[string]Node) {

	for key, value := range receivedNodeMap {
		_, ok := nodeMetaData.NodeMap[key]
		if !ok && key != nodeMetaData.Id {
			addNodeToMap(value, key)
			// TODO: Connect to the added node.
			connectToHelper(value.Addr)
		}
	}
}

func handleError(error error) {

}
