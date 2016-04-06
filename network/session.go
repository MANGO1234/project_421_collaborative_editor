package network

import (
	"github.com/satori/go.uuid"
	"net"
	"sync"
	"time"
)

type session struct {
	sync.WaitGroup
	id       string
	listener *net.TCPListener
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

func startNewSession(addr string) (*session, error) {
	lAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		return nil, err
	}
	//util.Debug("listening on ", lAddr.String())
	newSession := session{
		id:       uuid.NewV1().String(),
		listener: listener,
		done:     make(chan struct{}),
	}
	myNetMeta.Update(newSession.id, NodeMeta{addr, false})
	go newSession.listenForNewConn()
	go newSession.periodicallyReconnectDisconnectedNodes()
	go newSession.periodicallyCheckVersion()
	return &newSession, nil
}

func (s *session) end() {
	close(s.done)
	s.listener.Close()
	delta, _ := myNetMeta.Merge(NewQuitNetMeta(s.id, myAddr))
	Broadcast(createNetMetaUpdateMsg(delta))
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
		msg := createVersionCheckMsg()
		Broadcast(msg)
		time.Sleep(versionCheckInterval)
	}
}

func getLatestVersionVector() SerializableVersionVector {
	// TODO: this should retrieve the latest version vector for treedoc
	//       from somewhere
	return SerializableVersionVector{}
}

func createVersionCheckMsg() Message {
	latestMeta := myNetMeta.Copy()
	versionCheckMsgContent := VersionCheckMsgContent{
		latestMeta,
		getLatestVersionVector(),
	}
	content := versionCheckMsgContent.toJson()
	return NewSyncMessage(msgTypeVersionCheck, content)
}
