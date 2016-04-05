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
	"net"
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

// type ConnectMessage struct {
// 	Id          string
// 	Addr        string
// 	NetworkMeta NetMeta
// 	// TODO: maybe the treedoc and versionvector
// }

// this is just a stub for a version vector that can be marshelled
type SerializableVersionVector struct{}

type VersionCheckMsgContent struct {
	NetworkMeta   NetMeta
	VersionVector SerializableVersionVector
}

func (content *VersionCheckMsgContent) toJson() []byte {
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
var myNetMeta NetMeta
var myNetMetaRWMutex sync.RWMutex
var myConnectedNodes map[string]Node
var myDisconnectedNodes map[string]Node
var myConnectionMutex sync.Mutex
var mySession *session

func serveIncomingMessages(in <-chan Message) {
	for msg := range in {
		switch msg.Type {
		case msgTypeNetMetaUpdate:
			// TODO
		case msgTypeTreedocOp:
			// TODO
		case msgTypeVersionCheck:
			// TODO
		case msgTypeSync:
			// TODO
		default:
			// ignore and do nothing
		}
	}
}

func handleIncomingNetMeta(meta NetMeta) {

}

func serveBroadcastRequests(in <-chan Message) {
	for msg := range in {
		Broadcast(msg)
	}
}

func NewSyncMessage(msgType string, content []byte) Message {
	return Message{
		msgType,
		nil,
		content,
	}
}

func NewBroadcastMessage(msgType string, content []byte) Message {
	return Message{
		msgType,
		make(map[string]struct{}),
		content,
	}
}

func createNetMetaUpdateMsg(meta NetMeta) Message {
	return NewBroadcastMessage(msgTypeNetMetaUpdate, meta.ToJson())
}
