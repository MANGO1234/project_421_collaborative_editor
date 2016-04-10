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
	id            string
	addr          string
	msgChan       chan Message
	broadcastChan chan Message
	nodePool      *nodePool
	session       *session
}

var (
	ErrAlreadyConnected    = errors.New("network: already connected")
	ErrAlreadyDisconnected = errors.New("network: already disconnected")
)

// NewNetworkManager initiate a new NetworkManager with listening
// address addr to handle network operations
func NewNetworkManager(addr string) (*NetworkManager, error) {
	manager := NetworkManager{
		addr:          addr,
		msgChan:       make(chan Message, 10000),
		broadcastChan: make(chan Message, 10000),
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
			nm.handleIncomingNetMeta(msg)
		case msgTypeTreedocOp:
			nm.handleIncomingTreedocOp(msg)
		case msgTypeVersionCheck:
			nm.handleIncomingVersionCheck(msg)
		default:
			// ignore and do nothing
		}
	}
}

func (nm *NetworkManager) handleIncomingNetMeta(msg Message) {
	updates, err := newNetMetaFromJson(msg.Msg)
	if err != nil {
		return
	}
	newNodes, deltaNetMeta, changed := nm.nodePool.applyReceivedUpdates(updates)
	if changed {
		nm.session.handleNewNodes(newNodes)
		msg.Msg = deltaNetMeta.toJson()
		nm.BroadcastAsync(msg)
	}
}

func stubGetTreedocOpToPassOn(treedocOpPackage []byte) ([]byte, bool) {
	// TODO remove this
	return nil, false
}

func (nm *NetworkManager) handleIncomingTreedocOp(msg Message) {
	// get what should be broadcasted out from treedoc
	delta, changed := stubGetTreedocOpToPassOn(msg.Msg)
	if changed {
		msg.Msg = delta
		nm.BroadcastAsync(msg)
	}
}

func stubGetSyncInfoToReply(versionVector []byte) ([]byte, bool) {
	// TODO remove this
	return nil, false
}

func (nm *NetworkManager) handleIncomingVersionCheck(msg Message) {
	content, err := newVersionCheckMsgContentFromJson(msg.Msg)
	if err != nil {
		return
	}
	nm.handleIncomingNetMeta(newNetMetaUpdateMsg(nm.session.id, content.NetworkMeta))
}

func (nm *NetworkManager) serveBroadcastRequests() {
	for msg := range nm.broadcastChan {
		nm.session.broadcast(msg)
	}
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

// Broadcast msg asynchronously
func (nm *NetworkManager) BroadcastAsync(msg Message) {
	if nm.session != nil {
		go func() {
			nm.broadcastChan <- msg
		}()
	}
}

func (nm *NetworkManager) GetNetworkMetadata() string {
	return string(nm.nodePool.getLatestNetMetaJsonPrettyPrint())
}

func (nm *NetworkManager) GetNodePool() *nodePool {
	return nm.nodePool
}
