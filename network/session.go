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

//func (s *session) periodicallyReconnectDisconnectedNodes() {
//	s.Add(1)
//	defer s.Done()
//	for {
//		if s.ended() {
//			return
//		}
//		s.reconnectDisconnectedNodes()
//		time.Sleep(reconnectInterval)
//	}
//}
//
//func (s *session) reconnectDisconnectedNodes() {
//	nodes := s.manager.nodePool.getDisconnectedNodes()
//	for _, n := range nodes {
//		s.tryPokeOrConnect(n)
//	}
//}

func (s *session) broadcast(msg Message) {
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

//func (s *session) reconnectNode(n *node) {
//	n.stateMutex.Lock
//	if n.state == reconnecting
//}

func (s *session) handleDisconnectedNodes(nodes []*node) {
	for _, n := range nodes {
		n.stateMutex.Lock()
		if n.state == nodeStateDisconnected {
			go s.reconnectNodeThread(n)
		}
		n.stateMutex.Unlock()
	}
}

func (s *session) reconnectNodeThread(n *node) {

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
