package network

import (
	"github.com/satori/go.uuid"
	"net"
	"time"
)

// how often to check if version on two nodes match
const versionCheckInterval = 30 * time.Second

type session struct {
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
	newSession := session{
		id:       uuid.NewV1().String(),
		listener: listener,
		manager:  nm,
		done:     make(chan struct{}),
	}
	nm.nodePool.handleNewSession(&newSession)
	go newSession.listenForNewConn()
	go newSession.periodicallyCheckVersion()
	go newSession.serveIncomingMessages()
	nm.id = newSession.id
	nm.session = &newSession
	return nil
}

func (s *session) end() {
	close(s.done)
	s.listener.Close()
	s.manager.nodePool.handleEndSession(s)
}

func (s *session) ended() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (s *session) serveIncomingMessages() {
	for done := false; !done || len(s.manager.msgChan) > 0; {
		select {
		case msg := <-s.manager.msgChan:
			switch msg.Type {
			case MSG_TYPE_NET_META_UPDATE:
				s.handleIncomingNetMeta(msg)
			case MSG_TYPE_TREEDOC_OP:
				s.handleIncomingTreedocOp(msg)
			case MSG_TYPE_VERSION_CHECK:
				s.handleIncomingVersionCheck(msg)
			default:
				// ignore and do nothing
			}
		case <-s.done:
			done = true
		}
	}
}

func (s *session) handleIncomingNetMeta(msg Message) {
	updates, err := newNetMetaFromJson(msg.Msg)
	if err != nil {
		return
	}
	newNodes, deltaNetMeta, changed := s.manager.nodePool.applyReceivedUpdates(updates)
	if changed {
		s.handleNewNodes(newNodes)
		msg.Msg = deltaNetMeta.toJson()
		s.manager.nodePool.broadcast(msg)
	}
}

func (s *session) handleIncomingTreedocOp(msg Message) {
	if s.manager.TreeDocHandler != nil {
		s.manager.TreeDocHandler(msg.Msg)
	}
	s.manager.nodePool.broadcast(msg)
}

func stubGetSyncInfoToReply(versionVector []byte) ([]byte, bool) {
	// TODO remove this
	return nil, false
}

func (s *session) handleIncomingVersionCheck(msg Message) {
	content, err := newVersionCheckMsgContentFromJson(msg.Msg)
	if err != nil {
		return
	}
	s.handleIncomingNetMeta(newNetMetaUpdateMsg(s.id, content.NetworkMeta))
	syncInfo, shouldReply := stubGetSyncInfoToReply(content.VersionVector)
	if shouldReply {
		toSend := NewBroadcastMessage(s.id, MSG_TYPE_SYNC, syncInfo)
		go func() {
			s.manager.nodePool.sendMessageToNodeWithId(toSend, content.Source)
		}()
	}
}

// These functions launches major network threads
func (s *session) listenForNewConn() {
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

func (s *session) handleNewNodes(nodes []*node) {
	for _, n := range nodes {
		s.initiateNewNode(n)
	}
}

func (s *session) periodicallyCheckVersion() {
	for {
		time.Sleep(versionCheckInterval)
		msg := s.getLatestVersionCheckMsg()
		if s.ended() {
			return
		}
		s.manager.nodePool.broadcast(msg)
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
	return newSyncOrCheckMessage(MSG_TYPE_VERSION_CHECK, content)
}
