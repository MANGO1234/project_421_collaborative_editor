package network

import (
	"github.com/satori/go.uuid"
	"net"
	"sync"
	"time"
)

// how often to check version vector and network metadata in seconds
const (
	// how often to check if version on two nodes match
	versionCheckInterval = 30 * time.Second
	// how often to check whether it's time to (re)connect to disconnected nodes
	reconnectInterval = 30 * time.Second
)

type session struct {
	sync.WaitGroup
	id       string
	listener *net.TCPListener
	manager  *NetworkManager
	done     chan struct{}
}

func startNewSessionOnNetworkManager(nm *NetworkManager) error {
	lAddr, err := net.ResolveTCPAddr("tcp", nm.addr)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		return err
	}
	//util.Debug("listening on ", lAddr.String())
	newSession := session{
		id:       uuid.NewV1().String(),
		listener: listener,
		manager:  nm,
		done:     make(chan struct{}),
	}
	nm.nodePool.handleNewSession(&newSession)
	go newSession.listenForNewConn()
	go newSession.periodicallyCheckVersion()
	nm.id = newSession.id
	nm.session = &newSession
	return nil
}

func (s *session) end() {
	// TODO: we need to double check this
	// TODO: this isn't as graceful as it should be. We should wait
	// until all pending operations and broadcasts are performed
	close(s.done)
	s.listener.Close()
	delta := newQuitNetMeta(s.id, s.manager.addr)
	// TODO wait for all pending messages to be handled
	// TODO wait for all pending broadcast operations to complete
	s.broadcast(newNetMetaUpdateMsg(s.id, delta))
	s.disconnectAllConnectedNodes()
	s.Wait()
}

func (s *session) disconnectAllConnectedNodes() {
	// TODO
}

func (s *session) ended() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

// These functions launches major network threads
func (s *session) listenForNewConn() {
	s.Add(1)
	defer s.Done()
	for {
		conn, err := s.listener.Accept()
		if s.ended() {
			return
		}
		if err != nil {
			// TODO: What else can we do in this case
			continue
		}
		go s.handleNewConn(conn)
	}
}

func (s *session) broadcast(msg Message) {
	if msg.Visited == nil {

	} else {
		s.broadcastRecursive(msg)
	}
}

func (s *session) broadcastOnce(msg Message) {
	connected := s.manager.nodePool.getConnectedNodes()
	for _, n := range connected {
		s.sendMessageToNode(msg, n)
	}
}

func (s *session) broadcastRecursive(msg Message) {
	connected := s.manager.nodePool.getConnectedNodes()
	original := msg.Visited.copyVisitedNodes()
	msg.Visited.addAllFromNodeList(connected)
	for _, n := range connected {
		if !original.has(n.id) {
			succeeded := s.sendMessageToNode(msg, n)
			if !succeeded {
				delete(msg.Visited, n.id)
				// TODO initialize reconnect
			}
		}
	}
}

func (s *session) sendMessageToNodeWithId(msg Message, id string) bool {
	if n, ok := s.manager.nodePool.getNodeWithId(id); ok {
		return s.sendMessageToNode(msg, n)
	}
	return false
}

func (s *session) sendMessageToNode(msg Message, n *node) bool {
	msgJson := msg.toJson()
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.writeMutex.Lock()
	defer n.writeMutex.Unlock()
	if n.state == nodeStateConnected {
		err := n.writeMessageSlice(msgJson)
		if err != nil {
			n.close() // the reading thread handles reconnect
			return false
		}
		return true
	}
	return false
}

func (s *session) handleNewNodes(nodes []*node) {
	for _, n := range nodes {
		s.initiateNewNode(n)
	}
}

func (s *session) initiateNewNode(n *node) {
	if shouldConnect(s.manager.addr, n.addr) {
		go s.connectThread(n)
	} else {
		go s.pokeThread(n)
	}
}

func shouldConnect(localAddr, remoteAddr string) bool {
	return localAddr < remoteAddr
}

func (s *session) pokeThread(n *node) {
	// We poke when our addr is less than remote addr
	// This is because we try our best to deal with partial
	// connection in the user-initiated connect phase. A
	// user-initiated connect is considered partial if we
	// are not sure whether the other side has received our
	// information but we have obtained theirs.
	n.interval = 0
	for {
		// We don't poke any more if node is connected or if
		// a poke succeeded. The responsibility to connect is on
		// the other node, and we will listen for requests to
		// connect
		if n.state != nodeStateDisconnected || s.poke(n.id, n.addr) {
			return
		}
		n.interval += time.Second
		time.Sleep(n.interval)
	}
}

// returns whether the poke succeeded
func (s *session) poke(id, addr string) bool {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	defer conn.Close()
	n := newConnWrapper(conn)
	err = n.writeMessage(dialingTypePoke)
	if err != nil {
		return false
	}
	err = n.writeMessage(id)
	if err != nil {
		return false
	}
	match, err := n.readMessage()
	if err != nil {
		return false
	}
	if match == "true" {
		err = n.writeMessage(s.id)
		if err != nil {
			return false
		}
		err = n.writeMessage(s.manager.addr)
		if err != nil {
			return false
		}
		// TODO: not sure if the following is necessary when using tcp
		// but it gives more guarantees
		reply, err := n.readMessage()
		if err != nil {
			return false
		}
		return reply == "done"
	} else {
		s.manager.nodePool.forceNodeQuit(n)
		return true
	}
}

func (s *session) connectThread(n *node) {
	// We establish actual connection when our addr is greater
	// than remote addr.
	n.interval = 0
	for {
		if n.state != nodeStateDisconnected || s.connect(n) {
			return
		}
		n.interval += time.Second
		time.Sleep(n.interval)
	}
}

// returns whether connect succeeded
func (s *session) connect(n *node) bool {
	conn, err := net.Dial("tcp", n.addr)
	if err != nil {
		return false
	}
	n.setConn(conn)
	defer func(err error, n *node) {
		if err != nil {
			n.close()
			n.resetConn()
		}
	}(err, n)
	err = n.writeMessage(dialingTypeConnect)
	if err != nil {
		return false
	}
	err = n.writeMessage(n.id)
	if err != nil {
		return false
	}
	match, err := n.readMessage()
	if err != nil {
		return false
	}
	if match == "true" {
		err = n.writeMessage(s.id)
		if err != nil {
			return false
		}
		err = n.writeMessage(s.manager.addr)
		if err != nil {
			return false
		}
		err = n.sendMessage(s.getLatestVersionCheckMsg())
		if err != nil {
			return false
		}
		n.state = nodeStateConnected
		go s.sendThread(n.getSendWrapper())
		go s.receiveThread(n)
		return true
	} else {
		n.close()
		n.resetConn()
		s.manager.nodePool.forceNodeQuit(n)
		return true
	}
}

func (s *session) receiveThread(n *node) {
	for {
		if n.state == nodeStateQuitted {
			return
		}
		msg, err := n.receiveMessage()
		if err != nil {
			n.close()
			n.resetConn()
			n.state = nodeStateDisconnected
			if shouldConnect(s.manager.addr, n.addr) {
				go s.connectThread(n)
			}
		}
		s.manager.msgChan <- msg
	}
}

func (s *session) sendThread(sendWrapper *node) {
	for msg := range sendWrapper.outChan {
		err := sendWrapper.sendMessage(msg)
		if err != nil {
			sendWrapper.close()
			sendWrapper.outChan <- msg
			return
		}
	}
}

func (s *session) periodicallyCheckVersion() {
	s.Add(1)
	defer s.Done()
	for {
		time.Sleep(versionCheckInterval)
		if s.ended() {
			return
		}
		msg := s.getLatestVersionCheckMsg()
		s.broadcast(msg)
	}
}

func getLatestVersionVector() []byte {
	// TODO: this should retrieve the latest version vector for treedoc
	//       from somewhere
	return nil
}

func (s *session) getLatestVersionCheckMsg() Message {
	latestMeta := s.manager.nodePool.getLatestNetMetaCopy()
	versionCheckMsgContent := VersionCheckMsgContent{
		s.id,
		latestMeta,
		getLatestVersionVector(),
	}
	content := versionCheckMsgContent.toJson()
	return newSyncOrCheckMessage(msgTypeVersionCheck, content)
}
