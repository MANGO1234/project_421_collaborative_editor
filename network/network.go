// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

// TODO: debug mode to log any errors

package network

import (
	"encoding/json"
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

// NewNetworkManager initiate a new NetworkManager with listening
// address addr to handle network operations
func NewNetworkManager(addr string) (*NetworkManager, error) {
	manager := NetworkManager{
		addr:          addr,
		msgChan:       make(chan Message),
		broadcastChan: make(chan Message, 15),
		nodePool:      newNodePool(),
	}
	err := manager.startNewSession()
	if err != nil {
		return nil, err
	}
	go manager.serveBroadcastRequests()
	go manager.serveIncomingMessages()
	return &manager, nil
}

func (nm *NetworkManager) GetCurrentId() string {
	return nm.id
}

// ConnectTo registers the current node into the network of the node
// whose listening address is remoteAddr
func (nm *NetworkManager) ConnectTo(remoteAddr string) error {
	if nm.session == nil {
		err := nm.startNewSession()
		if err != nil {
			return err
		}
	}
	return nm.register(remoteAddr)
}

// Disconnect disconnects from the rest of the network voluntarily
func (nm *NetworkManager) Disconnect() error {
	if nm.session == nil {
		return errors.New("Already disconnected")
	}
	nm.session.end()
	nm.session = nil
	return nil
}

// Reconnect rejoins the network with new UUID.
func (nm *NetworkManager) Reconnect() error {
	if nm.session != nil {
		return errors.New("The node is already connected!")
	}
	return nm.startNewSession()
}

// Broadcast msg asynchronously
func (nm *NetworkManager) BroadcastAsync(msg Message) {
	go func() {
		nm.broadcastChan <- msg
	}()
}

func (nm *NetworkManager) broadcast(msg Message) {
	// TODO
}

func (nm *NetworkManager) GetNetworkMetadata() string {
	return string(nm.nodePool.getLatestNetMetaJsonPrettyPrint())
}

func (nm *NetworkManager) serveIncomingMessages() {
	for msg := range nm.msgChan {
		switch msg.Type {
		case msgTypeNetMetaUpdate:
			var delta NetMeta
			err := json.Unmarshal(msg.Msg, &delta)
			if err != nil {
				break
			}

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

func (nm *NetworkManager) serveBroadcastRequests() {
	for msg := range nm.broadcastChan {
		nm.broadcast(msg)
	}
}
