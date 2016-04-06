package network

import (
	"errors"
)

// Initialize initialize the network module, it's meant to be called once
// before calling any other function provided by this package
func Initialize(addr string) (string, error) {
	myAddr = addr
	myBroadcastChan = make(chan Message, 15)
	myMsgChan = make(chan Message)
	myNetMeta = NewNetMeta()
	myConnectedNodes = make(map[string]*node)
	myDisconnectedNodes = make(map[string]*node)
	go serveBroadcastRequests(myBroadcastChan)
	go serveIncomingMessages(myMsgChan)
	var err error
	mySession, err = startNewSession(addr)
	return mySession.id, err
}

// ConnectTo registers the current node into the network of the node
// whose listening address is remoteAddr
func ConnectTo(remoteAddr string) (id string, err error) {
	// start a new session if necessary
	if mySession == nil {
		mySession, err = startNewSession(myAddr)
		if err != nil {
			return
		}
	}
	id = mySession.id
	return id, register(myAddr, remoteAddr)
}

// Disconnect disconnects the current node from the network voluntarily
func Disconnect() error {
	if mySession == nil {
		return errors.New("Already disconnected")
	}
	mySession.end()
	mySession = nil
	return nil
}

// Re-initialize node with new UUID.
func Reconnect() (string, error) {
	if mySession != nil {
		return "", errors.New("The node is already connected!")
	}
	var err error
	mySession, err = startNewSession(myAddr)
	if err != nil {
		return "", err
	}
	return mySession.id, err
}

func BroadcastAsync(msg Message) {
	go func() {
		myBroadcastChan <- msg
	}()
}

func Broadcast(msg Message) {
	// TODO
}

func GetNetworkMetadata() string {
	return string(myNetMeta.ToJson())
}
