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
	versionCheckInterval = 30 * time.Second
	// how often to check whether it's time to (re)connect to disconnected nodes
	reconnectInterval = 30 * time.Second
)

// message types of messages sent to existing connection
const (
	msgTypeVersionCheck  = "versioncheck"
	msgTypeSync          = "sync"
	msgTypeNetMetaUpdate = "netmeta"
	msgTypeTreedocOp     = "treedocOp"
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
		time.Sleep(reconnectInterval)
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
		time.Sleep(versionCheckInterval)
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
