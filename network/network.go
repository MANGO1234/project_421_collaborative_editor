// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

// TODO: debug mode to log any errors

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
	// listening ip:port
	Addr string
	// quitted if the node voluntarily leaves the network
	// what's quitted stays quitted
	Quitted bool

	// fields below are for internal operations
	connected bool
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
	NodeMap map[string]*Node
}

// message type identifiers
const regMsg string = "registration"
const leaveMsg string = "disconnection"

// global variables
var myMeta NodeMetaData = NodeMetaData{NodeMap: make(map[string]*Node)} // keeps track of my NodeMetaData
var myAddr string
var myListener *net.TCPListener

// initialize local network listener
func Initialize(addr string) error {
	myAddr = addr
	return startNewSession(addr)
}

func startNewSession(addr string) error {
	lAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	myListener, err = net.ListenTCP("tcp", lAddr)
	if err == nil {
		fmt.Println("listening on ", lAddr.String())
		myMeta.Id = uuid.NewV1().String()
		myMeta.Addr = lAddr.String()
		go listenForConn(myListener)
	}
	return err
}

// listen for incoming connection to register
func listenForConn(listener *net.TCPListener) {
	for {
		conn, _ := listener.Accept()
		go handleConn(conn)
	}
}

// Disconnect from the network voluntarily
func Disconnect() {
	// refuse new incoming connections
	myListener.Close() // best attemp, error ignored
	// close all incoming and outgoing connection
	for _, node := range myMeta.NodeMap {
		if !node.Quitted {
			// close all existing incoming connections
			node.in.Connection.Close()
			// send disconnection notice
			err := node.out.ConnWriter.WriteMessage2(leaveMsg, make([]byte, 100))
			handleError(err)
			// close all outgoing connections
			node.out.Connection.Close()
			fmt.Println("disconnected --- ", node.Addr)
		}
	}
}

// Reconnect
func Reconnect() error {
	err := startNewSession(myAddr)
	if err != nil {
		return err
	}
	// TODO connect to all previously known nodes
	return err
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
	if msgInType == regMsg {
		msg, _ := json.Marshal(myMeta)
		msgWriter.WriteMessage2(regMsg, msg)
	}

	//add this node to nodeMap
	myMeta.Lock()
	_, ok := myMeta.NodeMap[msgIn.Id]
	if !ok {
		newNode := Node{Addr: msgIn.Addr, Quitted: false, connected: true}
		addNodeToMap(&newNode, msgIn.Id)
		// connect to this node
		connectToHelper(msgIn.Addr)
	}
	handleNewNodes(msgIn.NodeMap)
	myMeta.Unlock()
	for {
		continueRead(msgIn.Id, msgReader)
	}
}

func continueRead(id string, msgReader util.MessageReader) {
	//
	msgInType, _, err := msgReader.ReadMessage2()
	handleError(err)

	if msgInType == leaveMsg {
		result, ok := myMeta.NodeMap[id]
		if ok && result.connected {
			result.out.Connection.Close()
			result.Quitted = true
			result.connected = false
			fmt.Println(myMeta.NodeMap[id].Addr, "-----quitted")
		}
	}

}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	myMeta.Lock()
	defer myMeta.Unlock()
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
	msg, _ := json.Marshal(myMeta)

	err = msgWriter.WriteMessage2(regMsg, msg)
	handleError(err)

	// receive registration information
	msgType, msgBuff, err := msgReader.ReadMessage2()
	handleError(err)
	var newNodeData NodeMetaData
	if msgType == regMsg {
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

	addNodeToMap(&newNode, newNodeData.Id)

	handleNewNodes(newNodeData.NodeMap)

	return err
}

func GetNetworkMetadata() string {
	myMeta.Lock()
	defer myMeta.Unlock()
	meta, _ := json.Marshal(myMeta)
	return string(meta)
}

func Broadcast() {

}

// TODO restructure this
func addNodeToMap(nodeData *Node, nodeId string) {
	myMeta.NodeMap[nodeId] = nodeData
}

func handleNewNodes(receivedNodeMap map[string]*Node) {

	for key, value := range receivedNodeMap {
		_, ok := myMeta.NodeMap[key]
		if !ok && key != myMeta.Id {
			addNodeToMap(value, key)
			// TODO: Connect to the added node.
			connectToHelper(value.Addr)
		}
	}
}

func handleError(error error) {

}
