package network

import (
	"sync"
)

type node struct {
	id        string
	addr      string
	conn      *ConnWrapper
	countdown int // reconnect if this is 0 and node is disconnected
	interval  int // current interval to reconnect
	// TODO we might want to have a lock or even channel if we want
	//      to get more performance by multi-threading.
	//      We can have a buffered channel for broadcasting so that
	//      we can broadcast to multiple nodes at the same time
	//      Given the current organization of the code, this should
	//      be fairly easily acheived if we find the current
	//      performance unpromising
}

type nodepool struct {
	netmetaMutex     sync.RWMutex
	poolMutex        sync.Mutex
	netmeta          NetMeta
	connectedPool    map[string]*node
	disconnectedPool map[string]*node
}
