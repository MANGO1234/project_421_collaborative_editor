package network

import (
	"../util"
	"bufio"
	"encoding/json"
	"net"
	"sync"
)

const (
	uninitialized = iota
	stateConnected
	stateDisconnected
	stateQuitted
)

type NodeMeta struct {
	Addr    string
	Quitted bool
}

type NetMeta map[string]NodeMeta

// Design note:
// the mutex in the node also allows for more parallelism for performance
// although we are not takng full advantage of that at the moment
// It helps with broadcasting to different nodes at the same time
// We can also consider a buffered channel for sending purposes.

type node struct {
	sync.Mutex
	state     int
	id        string
	addr      string
	conn      net.Conn
	reader    *util.MessageReader
	writer    *util.MessageWriter
	countdown int // reconnect if this is 0 and node is disconnected
	interval  int // current interval to reconnect
}

type nodepool struct {
	netmetaMutex     sync.RWMutex
	poolMutex        sync.Mutex
	netmeta          map[string]NodeMeta
	connectedPool    map[string]*node
	disconnectedPool map[string]*node
}

func newNodeFromAddr(addr string) *node {
	var n node
	n.addr = addr
	return &n
}

func newNodeFromIdAddr(id, addr string) *node {
	n := newNodeFromAddr(addr)
	n.id = id
	return n
}

func newNodeFromConn(conn net.Conn) *node {
	var n node
	n.setConn(conn)
	return &n
}

func (n *node) setConn(conn net.Conn) {
	n.conn = conn
	n.reader = &util.MessageReader{bufio.NewReader(conn)}
	n.writer = &util.MessageWriter{bufio.NewWriter(conn)}
}

func (n *node) close() error {
	return n.conn.Close()
}

func (n *node) writeMessageSlice(msg []byte) error {
	return n.writer.WriteMessageSlice(msg)
}

func (n *node) writeMessage(msg string) error {
	return n.writer.WriteMessage(msg)
}

func (n *node) readMessage() (string, error) {
	return n.reader.ReadMessage()
}

func (n *node) ReadMessageSlice() ([]byte, error) {
	return n.reader.ReadMessageSlice()
}

func NewNetMeta() NetMeta {
	return make(map[string]NodeMeta)
}

func NewQuitNetMeta(id, addr string) NetMeta {
	netMeta := NewNetMeta()
	netMeta[id] = NodeMeta{addr, true}
	return netMeta
}

func NewJoinNetMeta(id, addr string) NetMeta {
	netMeta := NewNetMeta()
	netMeta[id] = NodeMeta{addr, false}
	return netMeta
}

func (netMeta NetMeta) Has(id string) bool {
	_, ok := netMeta[id]
	return ok
}

// return id, newNode, true if the update results in a change and "", nil, false otherwise
func (netMeta NetMeta) Update(id string, newNode NodeMeta) bool {
	if node, ok := netMeta[id]; !ok || (!node.Quitted && newNode.Quitted) {
		netMeta[id] = newNode
		return true
	}
	return false
}

// return the changes resulting from merge
func (netMeta NetMeta) Merge(netMeta2 NetMeta) (NetMeta, bool) {
	delta := NewNetMeta()
	for id, node := range netMeta2 {
		changed := netMeta.Update(id, node)
		if changed {
			delta[id] = node
		}
	}
	if len(delta) == 0 {
		return nil, false
	}
	return delta, true
}

func (netMeta NetMeta) ToJson() []byte {
	metaJson, _ := json.Marshal(netMeta)
	return metaJson
}

func (netMeta NetMeta) Copy() NetMeta {
	newMeta := NewNetMeta()
	for id, node := range netMeta {
		newMeta[id] = node
	}
	return newMeta
}
