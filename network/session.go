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

func (s *session) ended() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (nm *NetworkManager) startNewSession() error {
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
	go newSession.periodicallyReconnectDisconnectedNodes()
	go newSession.periodicallyCheckVersion()
	nm.id = newSession.id
	nm.session = &newSession
	return nil
}

func (s *session) end() {
	close(s.done)
	s.listener.Close()
	delta := newQuitNetMeta(s.id, s.manager.addr)
	s.manager.broadcast(newNetMetaUpdateMsg(delta))
	// TODO disconnect all connected nodes
	s.Wait()
}

// These functions launches major network threads
func (s *session) listenForNewConn() {
	s.Add(1)
	defer s.Done()
	for {
		if s.ended() {
			return
		}
		conn, err := s.listener.Accept()
		if err == nil {
			go s.handleNewConn(conn)
		}
	}
}

func (s *session) periodicallyReconnectDisconnectedNodes() {
	s.Add(1)
	defer s.Done()
	for {
		if s.ended() {
			return
		}
		reconnectDisconnectedNodes()
		time.Sleep(reconnectInterval)
	}
}

func reconnectDisconnectedNodes() {
	// TODO
}

func (s *session) periodicallyCheckVersion() {
	s.Add(1)
	defer s.Done()
	for {
		if s.ended() {
			return
		}
		msg := s.getLatestVersionCheckMsg()
		s.manager.broadcast(msg)
		time.Sleep(versionCheckInterval)
	}
}

func getLatestVersionVector() []byte {
	// TODO: this should retrieve the latest version vector for treedoc
	//       from somewhere
	return nil
}

func (s *session) getLatestVersionCheckMsg() Message {
	latestMeta := s.manager.nodePool.getLatestNetmetaCopy()
	versionCheckMsgContent := VersionCheckMsgContent{
		latestMeta,
		getLatestVersionVector(),
	}
	content := versionCheckMsgContent.toJson()
	return newSyncOrCheckMessage(msgTypeVersionCheck, content)
}
