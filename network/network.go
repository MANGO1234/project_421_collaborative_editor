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
	conn   net.Conn
	reader *util.MessageReader
	writer *util.MessageWriter
}

type NodeMetaData struct {
	sync.RWMutex //only for the map
	Id           string
	Addr         string
	NodeMap      map[string]*Node
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
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go handleConn(conn)
	}
}

// Disconnect from the network voluntarily
func Disconnect() {
	// refuse new incoming connections
	myListener.Close() // best attemp, error ignored
	// close all incoming and outgoing connection
	nodeList := getAllNodes()
	fmt.Println(nodeList)
	for _, node := range nodeList {
		if !node.Quitted {
			// close all existing incoming connections
			node.closeInConn()
			// send disconnection notice & disconnect
			if node.out != nil{
				err := node.out.writer.WriteMessage2(leaveMsg, make([]byte, 100))
				handleError(err)
				node.out.conn.Close()
				fmt.Println("disconnected --- ", node.Addr)
			}
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

// get a new ConnWrapper around a connection
func newConnWrapper(conn net.Conn) *ConnWrapper {
	msgWriter := util.MessageWriter{bufio.NewWriter(conn)}
	msgReader := util.MessageReader{bufio.NewReader(conn)}
	wrapper := ConnWrapper{conn, &msgReader, &msgWriter}
	return &wrapper
}

// handle node joining or rejoining
func handleConn(conn net.Conn) {
	// receive registration information
	wrapper := newConnWrapper(conn)

	msgInType, m, err := wrapper.reader.ReadMessage2()
	if err != nil || msgInType != regMsg {
		conn.Close()
		return
	}

	msgIn, err := jsonToMeta(m)
	if err != nil {
		conn.Close()
		return
	}

	fmt.Println("receiving-connection---: ", string(m))
	// write back registration information
	replyMsg := metaToJson()
	wrapper.writer.WriteMessage2(regMsg, replyMsg)

	//add this node to nodeMap
	node, ok := getNode(msgIn.Id)
	if !ok {
		newNode := Node{msgIn.Addr, false, true, wrapper, nil}
		putNewNode(&newNode, msgIn.Id)
		// connect to this node
		connectToHelper(msgIn.Addr)
	} else {
		if node.Quitted {
			conn.Close()
			return
		}
		if node.in != nil {
			node.in.conn.Close()
		}
		node.in = wrapper
	}
	handleNewNodes(msgIn.NodeMap)
	foreverRead(msgIn.Id, wrapper.reader)
}

func foreverRead(id string, msgReader *util.MessageReader) {
	for {
		msgInType, _, err := msgReader.ReadMessage2()
		if err != nil {
			handleError(err)
			return
		}
		switch msgInType {
		case leaveMsg:
			handleLeave(id)
			return
		default:
			return
		}

	}

}

func handleLeave (id string){
	node, ok := getNode(id)
	if ok && node.connected {
		node.in.conn.Close()
		node.out.conn.Close()
		node.Quitted = true
		node.connected = false
		fmt.Println(node.Addr, "-----quitted")
	}
}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	return connectToHelper(remoteAddr)
}

func connectToHelper(remoteAddr string) error {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	fmt.Println("---connecting--to---: \n", remoteAddr)
	wrapper := newConnWrapper(conn)

	// send registration information
	msg := metaToJson()
	err = wrapper.writer.WriteMessage2(regMsg, msg)
	handleError(err)

	// receive registration information
	msgType, msgBuff, err := wrapper.reader.ReadMessage2()
	if err != nil || msgType != regMsg {
		conn.Close()
		return err
	}
	receivedMeta, err := jsonToMeta(msgBuff)
	if err != nil {
		conn.Close()
		return err
	}

	// for checking
	fmt.Println("received reply type:", msgType)

	//add this node to nodeMap
	node, ok := getNode(receivedMeta.Id)
	if !ok {
		newNode := Node{receivedMeta.Addr, false, true, nil, wrapper}
		putNewNode(&newNode, receivedMeta.Id)
	} else {
		if node.Quitted {
			conn.Close()
			return nil
		}
		if node.out != nil {
			node.out.conn.Close()
		}
		node.out = wrapper
	}
	handleNewNodes(receivedMeta.NodeMap)

	return nil
}

func Broadcast() {

}

func handleNewNodes(receivedNodeMap map[string]*Node) {

	for key, value := range receivedNodeMap {
		_, ok := getNode(key)
		if !ok && key != myMeta.Id {
			putNewNode(value, key)
			connectToHelper(value.Addr)
		}
	}
}

func handleError(error error) {

}

func (node *Node) closeInConn() {
	if node.in != nil {
		node.in.conn.Close()
	}
}

func (node *Node) closeOutConn() {
	if node.out != nil {
		node.out.conn.Close()
	}
}

// Following functions are wrappers & helpers for accessing network metadata

func GetNetworkMetadata() string {
	myMeta.Lock()
	defer myMeta.Unlock()
	meta, _ := json.Marshal(myMeta)
	return string(meta)
}

func metaToJson() []byte {
	myMeta.RLock()
	defer myMeta.RUnlock()
	meta, _ := json.Marshal(myMeta)
	return meta
}

func jsonToMeta(msg []byte) (NodeMetaData, error) {
	var meta NodeMetaData
	err := json.Unmarshal(msg, &meta)
	return meta, err
}

func putNewNode(nodeData *Node, nodeId string) {
	myMeta.Lock()
	myMeta.NodeMap[nodeId] = nodeData
	myMeta.Unlock()
}

func getNode(nodeId string) (*Node, bool) {
	myMeta.RLock()
	result, ok := myMeta.NodeMap[nodeId]
	myMeta.RUnlock()
	return result, ok
}

// get a list of nodes of nodeMap
func getAllNodes() []*Node {
	myMeta.RLock()
	defer myMeta.RUnlock()
	n := len(myMeta.NodeMap)
	nodeList := make([]*Node, n)
	i := 0
	for _, node := range myMeta.NodeMap {
		nodeList[i] = node
		i++
	}
	return nodeList
}
