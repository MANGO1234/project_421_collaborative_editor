package network

// Initialize initialize the network module, it's meant to be called before
// calling any other function provided by this package
func Initialize(addr string) (string, error) {
	myAddr = addr
	myBroadcastChan = make(chan Message, 15)
	myMsgChan = make(chan Message)
	myNetMeta = netmeta.NewNetMeta()
	myConnectedNodes = make(map[string]Node)
	myDisconnectedNodes = make(map[string]Node)
	go serveBroadcastRequests(myBroadcastChan)
	go serveIncomingMessages(myMsgChan)
	return startNewSession(addr)
}

// ConnectTo registers the current node into the network of the node
// whose listening address is remoteAddr
func ConnectTo(remoteAddr string) (id string, err error) {
	// start a new session if necessary
	if mySession == nil {
		id, err = startNewSession()
		if err != nil {
			return
		}
	}
	id = mySession.id
	return id, register(remoteAddr)
}

// Disconnect disconnects the current node from the network voluntarily
func Disconnect() error {
	if mySession == nil {
		return errors.New("Already disconnected")
	}
	mySession.end()
	mySession = nil
}

// Re-initialize node with new UUID.
func Reconnect() (string, error) {
	if mySession != nil {
		return "", errors.New("The node is already connected!")
	}
	return startNewSession()
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
