// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

// TODO: debug mode to log any errors

package network

import (
	"errors"
	"net"
)

type NetworkManager struct {
	id             string
	addr           string
	msgChan        chan Message
	nodePool       *nodePool
	session        *session
	TreeDocHandler func([]byte)
}

var (
	ErrAlreadyConnected    = errors.New("network: already connected")
	ErrAlreadyDisconnected = errors.New("network: already disconnected")
)

// NewNetworkManager initiate a new NetworkManager with listening
// address addr to handle network operations
func NewNetworkManager(addr string) (*NetworkManager, error) {
	manager := NetworkManager{
		addr:     addr,
		msgChan:  make(chan Message, 30),
		nodePool: newNodePool(),
	}
	err := startNewSessionOnNetworkManager(&manager)
	if err != nil {
		return nil, err
	}
	return &manager, nil
}

func (nm *NetworkManager) GetCurrentId() string {
	return nm.id
}

// Notes on implementation of ConnectTo:
// To simplify the flow and design, we do not try to establish connection
// in the user command thread. Instead we put the load on the message handler
// which will in turn manage all connections

// ConnectTo registers the current node into the network of the node
// whose listening address is remoteAddr
func (nm *NetworkManager) ConnectTo(remoteAddr string) error {
	// TODO: maybe make the errors more friendly as it is user facing
	if nm.session == nil {
		err := startNewSessionOnNetworkManager(nm)
		if err != nil {
			return err
		}
	}
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	n := newConnWrapper(conn)
	err = n.writeMessage(dialingTypeRegister)
	if err != nil {
		return err
	}
	incomingNetMeta, err := n.readMessageSlice()
	if err != nil {
		return err
	}
	incoming, err := newNetMetaFromJson(incomingNetMeta)
	if err != nil {
		return err
	}
	defer func() { nm.msgChan <- newNetMetaUpdateMsg(nm.id, incoming) }()
	latestNetMeta := nm.nodePool.getLatestNetMetaJson()
	err = n.writeMessageSlice(latestNetMeta)
	if err != nil {
		return errors.New("Partially connected: unable to send message to " +
			"requested node, but was able to receive information.")
	}
	return nil
}

// Disconnect disconnects from the rest of the network voluntarily
func (nm *NetworkManager) Disconnect() error {
	if nm.session == nil {
		return ErrAlreadyDisconnected
	}
	nm.session.end()
	nm.session = nil
	return nil
}

// Reconnect rejoins the network with new UUID.
func (nm *NetworkManager) Reconnect() error {
	if nm.session != nil {
		return ErrAlreadyConnected
	}
	return startNewSessionOnNetworkManager(nm)
}

// Broadcast msg asynchronously return whether the msg
// if session has ended it will not broadcast
func (nm *NetworkManager) Broadcast(msg Message) {
	s := nm.session // this is necessary for thread safety and to avoid nil pointer dereference
	if s != nil || !s.ended() {
		nm.nodePool.broadcast(msg)
	}
}

// Send msg to a node with specified id
// if session has ended it will not send
func (nm *NetworkManager) SendMessageToNodeWithId(msg Message, id string) {
	s := nm.session // this is necessary for thread safety and to avoid nil pointer dereference
	if s != nil || !s.ended() {
		nm.nodePool.sendMessageToNodeWithId(msg, id)
	}
}

func (nm *NetworkManager) GetNetworkMetadata() string {
	return string(nm.nodePool.getLatestNetMetaJsonPrettyPrint())
}

func (nm *NetworkManager) GetNodePool() *nodePool {
	return nm.nodePool
}

func (nm *NetworkManager) SetTreeDocHandler(fn func([]byte)) {
	nm.TreeDocHandler = fn
}

func (nm *NetworkManager) RemoveTreeDocHandler() {
	nm.TreeDocHandler = nil
}
