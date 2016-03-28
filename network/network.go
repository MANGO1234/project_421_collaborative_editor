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

type NodeMetaDataUpdate struct {
	NodeMeta     NodeMetaData
	VisitedNodes map[string]bool
}

// message type identifiers
const regMsg string = "registration"
const leaveMsg string = "disconnection"
const metaUpdateMsg string = "metaUpdate"

// global variables
var myMeta NodeMetaData = NodeMetaData{NodeMap: make(map[string]*Node)} // keeps track of my NodeMetaData

// initialize local network listener
func Initialize(addr string) (id string, err error) {
	return startNewSession(addr)
}

func startNewSession(addr string) (id string, err error) {
	if myMeta.initialized {
		return
	}
	lAddr, err := net.ResolveTCPAddr("tcp", addr)
	myListener, err := net.ListenTCP("tcp", lAddr)
	if err == nil {
		fmt.Println("listening on ", lAddr.String())
		myMeta.Id = uuid.NewV1().String()
		id = myMeta.Id
		myMeta.Addr = lAddr.String()
		myMeta.listener = myListener
		myMeta.initialized = true
		go listenForConn(myListener)
	}
	return
}

// listen for incoming connection to register
func listenForConn(listener *net.TCPListener) {
	for {
		conn, err := listener.Accept()
		if err == nil {
			go handleConn(conn)
		}
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
				node.closeOutConn()
				fmt.Printf("disconnect from: %s\n", node.Addr)
			}
		}
	}
	myMeta.initialized = false
}

// Re-initialize node with new UUID.
func Reconnect() error {
	if myMeta.initialized == true {
		return errors.New("Please use reconnect only after you disconnect.")
	}
	_, err := startNewSession(myMeta.Addr)
	if err == nil {
		// connect to all previously known nodes
		connectKnownNodes()
	}
	return err
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
	changed, visitedNodes := handleNodesMap(msgIn.NodeMap)
	visitedNodes[msgIn.Id] = true
	visitedNodes[myMeta.Id] = true

	// broadcast new nodeMap to peers only if current call has resulted change in current nodeMap
	if changed || !ok {
		broadcastToPeer(visitedNodes)
	}
	foreverRead(msgIn.Id, wrapper.reader)
}

// handle
func foreverRead(id string, msgReader *util.MessageReader) {
	for {
		msgInType, msg, err := msgReader.ReadMessage2()
		if err != nil {
			handleError(err)
			return
		}
		switch msgInType {
		case leaveMsg:
			handleLeave(id)
			return
		case metaUpdateMsg:
			receivedUpdate, err := JsonToMetaUpdate(msg)
			if err != nil {
				handleMetaUpdate(receivedUpdate)
				return
			}
		default:
			return
		}

	}

}

func handleMetaUpdate(receivedUpdate NodeMetaDataUpdate) {
	changed, contactedNodes := handleNodesMap(receivedUpdate.NodeMeta.NodeMap)
	if changed {
		for key, _ := range receivedUpdate.VisitedNodes {
			contactedNodes[key] = true
		}
		contactedNodes[myMeta.Id] = true
		broadcastToPeer(contactedNodes)
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
	visitedNodes := make(map[string]bool)
	visitedNodes[id] = true
	visitedNodes[myMeta.Id] = true

	broadcastToPeer(visitedNodes)
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
	//	handleNodesMap(receivedMeta.NodeMap)

	return nil
}

func Broadcast() {

}

// broadcast current node meta data to all peers except for those who have been contacted already
func broadcastToPeer(visitedNodes map[string]bool) {
	nodeMap := getNodeMap()
	metaUpdate := getMetaUpdateJson(visitedNodes)
	for id, node := range nodeMap {
		if !node.Quitted && !visitedNodes[id] {
			node.sendMsg(metaUpdateMsg, metaUpdate)
			fmt.Println("broadcasted change to " + node.Addr)
		}
	}
}

// handle received nodeMap appropriately
// return a bool indicating if receivedNodeMap has caused any change to current metadata
// return a list of nodeIds being contacted during this process
func handleNodesMap(receivedNodeMap map[string]*Node) (bool, map[string]bool) {
	var changed bool
	contactedNodes := make(map[string]bool)

	for key, value := range receivedNodeMap {
		node, ok := getNode(key)
		if !ok && key != myMeta.Id && value.Addr != myMeta.Addr {
			putNewNode(value, key)
			if !value.Quitted {
				connectToHelper(value.Addr)
				contactedNodes[key] = true
			}
			changed = true
		}

		if ok && value.Quitted && !node.Quitted {
			handleLeave(key)
			changed = true
		}
	}

	return changed, contactedNodes
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

func (node *Node) sendMsg(msgType string, msg []byte) error {
	if node.out != nil {
		return node.out.writer.WriteMessage2(msgType, msg)
	}
	return errors.New("No out connection available.")
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

// get a copy of nodeMap
func getNodeMap() map[string]*Node {
	myMeta.RLock()
	defer myMeta.RUnlock()
	mapCopy := make(map[string]*Node)
	for key, value := range myMeta.NodeMap {
		mapCopy[key] = value
	}

	return mapCopy
}

func getMetaUpdateJson(visitedNodes map[string]bool) []byte {
	myMeta.RLock()
	defer myMeta.RUnlock()
	newMetaUpdate := NodeMetaDataUpdate{myMeta, visitedNodes}
	metaUpdate, _ := json.Marshal(newMetaUpdate)
	return metaUpdate
}

func JsonToMetaUpdate(message []byte) (NodeMetaDataUpdate, error) {
	var metaUpdate NodeMetaDataUpdate
	err := json.Unmarshal(message, &metaUpdate)
	return metaUpdate, err
}
