package network

import (
	"../util"
	"net"
	"sync"
)

// Design note:
// the mutex in the node also allows for more parallelism for performance
// although we are not taking full advantage of that at the moment
// It helps with broadcasting to different nodes at the same time
// We can also consider a buffered channel for sending purposes.

// node state
const (
	nodeStateUninitialized = iota
	nodeStateConnected
	nodeStateDisconnected
	nodeStateLeft
)

type node struct {
	sync.Mutex
	state     int
	id        string
	addr      string
	conn      net.Conn
	reader    *util.MessageReader
	writer    *util.MessageWriter
	interval  int // current interval to reconnect
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
