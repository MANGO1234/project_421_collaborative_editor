// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

// TODO: debug mode to log any errors

package network

import (
	"../util"
	"./netmeta"
	"encoding/json"
	"errors"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

type Node struct {
	addr string
	conn *ConnWrapper
}

// get a new ConnWrapper around a connection
func newConnWrapper(conn net.Conn) *ConnWrapper {
	msgWriter := util.MessageWriter{bufio.NewWriter(conn)}
	msgReader := util.MessageReader{bufio.NewReader(conn)}
	wrapper := ConnWrapper{conn, &msgReader, &msgWriter}
	return &wrapper
}

type Message struct {
	Type    string
	Visited map[string]struct{}
	Msg     []byte
}

type ConnWrapper struct {
	conn   net.Conn
	reader *util.MessageReader
	writer *util.MessageWriter
}

type session struct {
	sync.WaitGroup
	id       string
	listener *net.TCPListener
	done     chan struct{}
}

// type ConnectMessage struct {
// 	Id          string
// 	Addr        string
// 	NetworkMeta netmeta.NetMeta
// 	// TODO: maybe the treedoc and versionvector
// }

// this is just a stub for a version vector that can be marshelled
type SerializableVersionVector struct{}

type VersionCheckMsgContent struct {
	NetworkMeta   netmeta.NetMeta
	VersionVector SerializableVersionVector
}

func (content *versionCheckMsgContent) toJson() []byte {
	contentJson, _ := json.Marshal(content)
	return contentJson
}

// constants

// how often to check version vector and network metadata in seconds
const (
	// how often to check if version on two nodes match
	versionCheckInterval = 30
	// how often to check whether it's time to (re)connect to disconnected nodes
	reconnectInterval = 30
)

// message types of messages sent to existing connection
const (
	msgTypeVersionCheck  = "versioncheck"
	msgTypeSync          = "sync"
	msgTypeNetMetaUpdate = "netmeta"
	msgTypeTreedocOp     = "treedocOp"
)

// the purpose of dialing to a node
const (
	// client-initiated call to connect to a remote node
	dialingTypeRegister = "register"
	// poke a known node so it has information to connect
	dialingTypePoke = "poke"
	// establish persistent connection between the nodes
	dialingTypeEstablishConnection = "connect"
)

// states to keep track of
var myAddr string
var myMsgChan chan Message
var myBroadcastChan chan Message
var myNetMeta netmeta.NetMeta
var myNetMetaRWMutex sync.RWMutex
var myConnectedNodes map[string]Node
var myDisconnectedNodes map[string]Node
var myConnectionMutex sync.Mutex
var mySession *session

// initialize local network listener
func Initialize(addr string) (string, error) {
	myAddr = addr
	myBroadcastChan = make(chan Message, 15)
	myMsgChan = make(chan Message)
	myNetMeta = netmeta.NewNetMeta()
	myConnectedNodes = make(map[string]Node)
	myDisconnectedNodes = make(map[string]Node)
	go serveBroadcastRequests(myBroadcastChan)
	go serveIncomingMessages(myMsgChan)
	return startNewSession(addr)
}

func (s *session) ended() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func getNewSession(listener *net.TCPListener) *session {
	s := session{uuid.NewV1().String(), listener, make(chan struct{})}
	myNetMeta.Update(s.id, netmeta.NodeMeta{myAddr, false})
	go s.listenForNewConn()
	go s.periodicallyReconnectDisconnectedNodes()
	go s.periodicallyCheckVersion()
	return s
}

func (s *session) end() {
	close(s.done)
	s.listener.Close()
	delta := myNetMeta.Update(s.id, netmeta.NodeMeta{myAddr, true})
	Broadcast(createNetMetaUpdateMsg(delta))
	// TODO disconnect all connected nodes
	s.wg.Wait()
}

func startNewSession() (string, error) {
	lAddr, err := net.ResolveTCPAddr("tcp", myAddr)
	if err != nil {
		return "", err
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		return "", err
	}
	//util.Debug("listening on ", lAddr.String())
	mySession = getNewSession(listener)
	return mySession.id, nil
}

// These functions launches major network threads
func (s *session) listenForNewConn() {
	s.Add(1)
	defer s.Done()
	for {
		if s.ended() {
			return
		}
		conn, err := s.listener.Accept()
		if err == nil {
			go handleNewConn(conn)
		}
	}
}

func handleIncomingNetMeta(meta netmeta.NetMeta) {

}

// TODO: not too sure how to organize yet
//       might want to have locks here or maybe in netmeta
func getLatestMeta() []byte {
	return myNetMeta.toJson()
}

func handleNewConn(conn net.Conn) {
	defer conn.Close()
	wrapper := newConnWrapper(conn)
	// distinguish purpose of this connection
	purpose, err := wrapper.reader.ReadMessage()
	if err != nil {
		return
	}
	switch purpose {
	case dialingTypePoke:
		expectedId, err := wrapper.reader.ReadMessage()
		if err != nil {
			return
		}
		if expectedId == mySession.id {
			err = wrapper.writer.WriteMessageSlice("true")
			if err != nil {
				return
			}
		} else {
			wrapper.writer.WriteMessageSlice("false")
			return
		}
		fallthrough
	case dialingTypeRegister:
		// send latest netmeta to the connecting node
		latestMeta := getLatestMeta()
		err = wrapper.writer.WriteMessageSlice(latestMeta)
		if err != nil {
			return
		}
		// retrieve the latest netmeta from the connecting node
		msg, err := wrapper.reader.ReadMessageSlice()
		if err != nil {
			return
		}
		conn.Close()
		var incomingMeta netmeta.NetMeta
		err = json.Unmarshal(registrationJson, &incomingMeta)
		if err != nil {
			return
		}
		// establish persistent connection and perform any necessary broadcast
		handleIncomingNetMeta(incomingMeta)
	case dialingTypeEstablishConnection:
		// TODO: actually accept persistent
	default:
		// invalid purpose
		return
	}
}

// All the following functions assume an Initialize call has been made
func ConnectTo(remoteAddr string) (id string, err error) {
	// start a new session if necessary
	if mySession == nil {
		id, err = startNewSession()
		if err != nil {
			return
		}
	}
	id = mySession.id
	return id, register(remoteAddr)
}

func register(remoteAddr string) error {
	return registerOrPokeHelper("", remoteAddr)
}

// this method doesn't try to establish a persistent connection
// it's goal is to register into the remote network and communicate
// the netmeta state between the two networks
// The actual persisting connection is to be established later
//
// Note on the design:
// This design avoids infinite reconnection when a connection can be
// established and avoids deadlock
//
// For example: when A and B connect to each other simutaneously,
// with a deterministic algorithm, it's possible that the two nodes
// performs symmetric actions, causing connections to be constantly
// replaced and failure to agree on one single connection between
// the nodes. If we try to use locks or channels to forbid this
// unnecessary reconnection loop, it's easy to get deadlock. Having
// two connections between the nodes solves this problem in a way
// but handling two connections when we only need to handle one is
// costly and leads to other problems associated with handling more
// connections
func registerOrPokeHelper(id, remoteAddr string) error {
	// connect to remote node
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	wrapper := newConnWrapper(conn)
	// indicate intention of this dial
	var intention string
	if id == "" {
		intention = dialingTypeRegister
	} else {
		intention = dialingTypePoke
	}
	err = wrapper.writer.WriteMessage(intention)
	if err != nil {
		return err
	}
	// when poking, we need to make sure the other node has the expected id
	// if it doesn't we should treat the node associated the id as quitted
	if intention == dialingTypePoke {
		err = wrapper.writer.WriteMessage(id)
		if err != nil {
			return err
		}
		match, err := wrapper.reader.ReadMessage()
		if err != nil {
			return err
		}
		if match == "false" {
			// TODO: mark node as deleted
			return nil
		}
	}
	// retrieve latest netmeta from remote node
	msg, err := wrapper.reader.ReadMessageSlice()
	if err != nil {
		return err
	}
	var incomingMeta netmeta.NetMeta
	err = json.Unmarshal(registrationJson, &incomingMeta)
	if err != nil {
		return err
	}
	// send latest netmeta to the remote node
	latestMeta := getLatestMeta()
	wrapper.writer.WriteMessageSlice(latestMeta)
	// The write is this node's best attempt, with the netmeta received, this node
	// considers itself to be part of the network; Thus no error handling
	conn.Close()
	// establish persistent connection and perform any necessary broadcast
	handleIncomingNetMeta(incomingMeta)
}

func (node *Node) reconnect() {
	// TODO reconnect
}

func (s *session) periodicallyReconnectDisconnectedNodes() {
	s.Add(1)
	defer s.Done()
	for {
		if s.ended() {
			return
		}
		reconnectDisconnectedNodes()
		time.Sleep(time.Second * reconnectInterval)
	}
}

func reconnectDisconnectedNodes() {
	// TODO
}

func (s *session) periodicallyCheckVersion() {
	s.Add(1)
	defer s.Done()
	for {
		if s.ended() {
			return
		}
		msg := createVersionCheckMsg()
		Broadcast(msg)
		time.Sleep(time.Second * versionCheckInterval)
	}
}

func getLatestVersionVector() SerializableVersionVector {
	// TODO: this should retrieve the latest version vector for treedoc
	//       from somewhere
	return nil
}

func createVersionCheckMsg() Message {
	latestMeta := myNetMeta.Copy()
	versionCheckMsgContent := VersionCheckMsgContent{
		latestMeta,
		getLatestVersionVector(),
	}
	content := versionCheckMsgContent.toJson()
	return NewSyncMessage(msgTypeVersionCheck, content)
}

func serveIncomingMessages(in <-chan Message) {
	for msg := range in {
		//TODO
	}
}

func serveBroadcastRequests(in <-chan Message) {
	for msg := range in {
		Broadcast(msg)
	}
}

func NewSyncMessage(msgType string, content []byte) {
	return Message{
		msgType,
		nil,
		content,
	}
}

func NewBroadcastMessage(msgType string, content []byte) {
	return Message{
		msgType,
		make(map[string]struct{}),
		content,
	}
}

func createNetMetaUpdateMsg(meta netmeta.NetMeta) Message {
	return NewBroadcastMessage(msgTypeNetMetaUpdate, meta.ToJson())
}

// Disconnect from the network voluntarily
func Disconnect() error {
	if mySession == nil {
		return errors.New("Already disconnected")
	}
	mySession.end()
	mySession = nil
}

// Re-initialize node with new UUID.
func Reconnect() (string, error) {
	if mySession != nil {
		return "", errors.New("The node is already connected!")
	}
	return startNewSession()
}

func BroadcastAsync(msg Message) {
	go func() {
		myBroadcastChan <- msg
	}()
}

func Broadcast(msg Message) {
	// TODO
}

func GetNetworkMetadata() string {
	return string(myNetMeta.ToJson())
}
