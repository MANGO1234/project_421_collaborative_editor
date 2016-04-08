// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

// TODO: debug mode to log any errors

package network

import (
	"errors"
)

type NetworkManager struct {
	id            string
	addr          string
	msgChan       chan Message
	broadcastChan chan Message
	nodePool      *nodePool
	session       *session
}

var (
	ErrAlreadyConnected = errors.New("network: already connected")
	ErrAlreadyDisonnected = errors.New("network: already disconnected")
)

// NewNetworkManager initiate a new NetworkManager with listening
// address addr to handle network operations
func NewNetworkManager(addr string) (*NetworkManager, error) {
	manager := NetworkManager{
		addr:          addr,
		msgChan:       make(chan Message),
		broadcastChan: make(chan Message, 15),
		nodePool:      newNodePool(),
	}
	err := startNewSessionOnNetworkManager(&manager)
	if err != nil {
		return nil, err
	}
	go manager.serveBroadcastRequests()
	go manager.serveIncomingMessages()
	return &manager, nil
}

func (nm *NetworkManager) serveIncomingMessages() {
	for msg := range nm.msgChan {
		switch msg.Type {
		case msgTypeNetMetaUpdate:
			nm.handleIncomingNetMeta(msg.Msg)
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

func TODO(sth interface{}) {

}

func (nm *NetworkManager) handleIncomingNetMeta(incoming []byte) error {
	incomingNetMeta, err := newNetMetaFromJson(incoming)
	if err != nil {
		return err
	}
	TODO(incomingNetMeta)
	return nil
	// TODO handle it
}

func (nm *NetworkManager) serveBroadcastRequests() {
	for msg := range nm.broadcastChan {
		nm.broadcast(msg)
	}
}

func (nm *NetworkManager) GetCurrentId() string {
	return nm.id
}

// ConnectTo registers the current node into the network of the node
// whose listening address is remoteAddr
func (nm *NetworkManager) ConnectTo(remoteAddr string) error {
	if nm.session == nil {
		err := startNewSessionOnNetworkManager(nm)
		if err != nil {
			return err
		}
	}
	return nm.register(remoteAddr)
}

// Disconnect disconnects from the rest of the network voluntarily
func (nm *NetworkManager) Disconnect() error {
	if nm.session == nil {
		return ErrAlreadyDisonnected
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

// Broadcast msg asynchronously
func (nm *NetworkManager) BroadcastAsync(msg Message) {
	go func() {
		nm.broadcastChan <- msg
	}()
}

func (nm *NetworkManager) broadcast(msg Message) {

}

func (nm *NetworkManager) GetNetworkMetadata() string {
	return string(nm.nodePool.getLatestNetMetaJsonPrettyPrint())
}
