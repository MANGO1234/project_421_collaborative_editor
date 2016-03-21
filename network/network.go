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
	"errors"
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
	listener     *net.TCPListener
	initialized  bool
}

// message type identifiers
const regMsg string = "registration"
const leaveMsg string = "disconnection"

// global variables
var myMeta NodeMetaData = NodeMetaData{NodeMap: make(map[string]*Node)} // keeps track of my NodeMetaData

// initialize local network listener
func Initialize(addr string) error {
	return startNewSession(addr)
}

func startNewSession(addr string) error {
	if myMeta.initialized {
		return nil
	}
	lAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	myListener, err := net.ListenTCP("tcp", lAddr)
	if err == nil {
		fmt.Println("listening on ", lAddr.String())
		myMeta.Id = uuid.NewV1().String()
		myMeta.Addr = lAddr.String()
		myMeta.listener = myListener
		myMeta.initialized = true
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
	myMeta.listener.Close() // best attemp, error ignored
	// close all incoming and outgoing connection
	nodeList := getAllNodes()
	for _, node := range nodeList {
		if !node.Quitted {
			// close all existing incoming connections
			node.closeInConn()
			// send disconnection notice & disconnect
			if node.out != nil {
				err := node.out.writer.WriteMessage2(leaveMsg, make([]byte, 100))
				handleError(err)
				node.out.conn.Close()
				fmt.Printf("disconnect from: %s\n", node.Addr)
			}
		}
	}
	myMeta.initialized = false
}

// Reconnect
func Reconnect() error {
	if myMeta.initialized == true {
		return errors.New("No reconnection needed.")
	}
	err := startNewSession(myMeta.Addr)
	if err != nil {
		return err
	}
	// connect to all previously known nodes
	connectKnownNodes()
	return nil
}

func connectKnownNodes() {
	nodeList := getAllNodes()
	for _, node := range nodeList {
		if !node.Quitted {
			err := connectToHelper(node.Addr)
			// TODO: handle connection error here
			if err != nil {

			}
		}
	}
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

	// write back registration information
	replyMsg := metaToJson()
	wrapper.writer.WriteMessage2(regMsg, replyMsg)

	//add this node to nodeMap
	node, ok := getNode(msgIn.Id)
	if !ok {
		newNode := Node{msgIn.Addr, false, true, wrapper, nil}
		putNewNode(&newNode, msgIn.Id)
		fmt.Printf("received connection: %s(%s)\n", msgIn.Id, msgIn.Addr)
		connectToHelper(msgIn.Addr)
	} else {
		if node.Quitted {
			conn.Close()
			node.closeInConn()
			node.closeOutConn()
			return
		}
		node.closeInConn()
		node.in = wrapper
		node.connected = true
		fmt.Printf("received connection: %s(%s)\n", msgIn.Id, msgIn.Addr)
	}
	handleNewNodes(msgIn.NodeMap)
	//broadcastToPeer(newNodeMsg)
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

func handleLeave(id string) {
	node, ok := getNode(id)
	if ok {
		node.closeInConn()
		node.closeOutConn()
		node.Quitted = true
		node.connected = false
		fmt.Printf("node quitted: %s(%s)\n", id, node.Addr)
	}
}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	if myMeta.initialized == false {
		err := Reconnect()
		if err != nil {
			return err
		}
	}

	if myMeta.Addr == remoteAddr {
		return errors.New("Please enter an address that's different from your local address.")
	}

	// TODO : fix possible connect to an address twice problem maybe by reordering method call
	return connectToHelper(remoteAddr)
}

// TODO: possible refactoring?
func startConnection(remoteAddr string, nodeId string) {

}

func connectToHelper(remoteAddr string) error {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	//	fmt.Println("connecting to:", remoteAddr)
	wrapper := newConnWrapper(conn)

	// send registration information
	msg := metaToJson()
	err = wrapper.writer.WriteMessage2(regMsg, msg)
	if err != nil {
		conn.Close()
		return err
	}

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

	//add this node to nodeMap
	node, ok := getNode(receivedMeta.Id)
	if !ok {
		newNode := Node{receivedMeta.Addr, false, true, nil, wrapper}
		putNewNode(&newNode, receivedMeta.Id)
	} else {
		if node.Quitted {
			conn.Close()
			node.closeInConn()
			node.closeOutConn()
			return nil
		}
		node.closeOutConn()
		node.out = wrapper
		node.connected = true
	}
	fmt.Printf("dialed connection: %s(%s)\n", receivedMeta.Id, receivedMeta.Addr)
	//	handleNewNodes(receivedMeta.NodeMap)

	return nil
}

func Broadcast() {

}

func handleNewNodes(receivedNodeMap map[string]*Node) {

	for key, value := range receivedNodeMap {
		node, ok := getNode(key)
		if !ok && key != myMeta.Id && value.Addr != myMeta.Addr {
			putNewNode(value, key)
			if !value.Quitted {
				connectToHelper(value.Addr)
			}
		}

		if ok && value.Quitted && !node.Quitted {
			handleLeave(key)
		}
	}
}

func handleError(error error) {

}

func (node *Node) closeInConn() {
	if node.in != nil {
		node.in.conn.Close()
		node.in = nil
	}
}

func (node *Node) closeOutConn() {
	if node.out != nil {
		node.out.conn.Close()
		node.out = nil
	}
}

func (node *Node) closeAllConn() {
	node.closeInConn()
	node.closeOutConn()
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
