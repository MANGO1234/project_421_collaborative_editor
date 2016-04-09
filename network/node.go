package network

import (
	"../util"
	"net"
	"sync"
	"time"
)

// Design note:
// the mutex in the node also allows for more parallelism for performance
// although we are not taking full advantage of that at the moment
// It helps with broadcasting to different nodes at the same time
// We can also consider a buffered channel for sending purposes.

// node state
const (
	nodeStateDisconnected = iota
	nodeStateReconnecting
	nodeStateConnected
	nodeStateQuitted
)

type node struct {
	stateMutex sync.Mutex
	writeMutex sync.Mutex
	state      int
	id         string
	addr       string
	outChan    chan Message
	conn       net.Conn
	reader     *util.MessageReader
	writer     *util.MessageWriter
	interval   time.Duration // current interval to reconnect
}

func newNodeFromAddr(addr string) *node {
	var n node
	n.addr = addr
	return &n
}

func newNodeFromIdAddr(id, addr string) *node {
	n := newNodeFromAddr(addr)
	n.id = id
	n.outChan = make(chan Message, 30)
	return n
}

func newConnWrapper(conn net.Conn) *node {
	var n node
	n.setConn(conn)
	return &n
}

// pre-condition: nodeMeta.Left == false
func newNodeFromIdNodeMeta(id string, nodeMeta NodeMeta) *node {
	n := newNodeFromIdAddr(id, nodeMeta.Addr)
	n.state = nodeStateDisconnected
	return n
}

func (n *node) getSendWrapper() *node {
	return &node{
		conn:    n.conn,
		reader:  n.reader,
		writer:  n.writer,
		outChan: n.outChan,
	}
}

func (n *node) leave() {
	// TODO: figure out locking
	n.close()
}

// if err occurs, close conn, set state to disconnect
// return error so the caller can decide whether to do any other
// error handling
func (n *node) handleAndReturnError(err error) error {
	if err != nil {
		// TODO: test and see if the commented out code is working as expected
		//       ie. should not make a legitimate node leave
		//
		//if _, ok := err.(net.Error); !ok {
		//	// if it's not a network error, we assume the node to be an imposter
		//	// and force it to leave the system
		//	n.state = nodeStateLeft
		//}
		n.close()
	}
	return err
}
