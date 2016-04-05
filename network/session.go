package network

import (
	"../util"
	"./netmeta"
	"encoding/json"
	"errors"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

func (s *session) ended() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func getNewSession(listener *net.TCPListener) *session {
	s := session{uuid.NewV1().String(), listener, make(chan struct{})}
	myNetMeta.Update(s.id, netmeta.NodeMeta{myAddr, false})
	go s.listenForNewConn()
	go s.periodicallyReconnectDisconnectedNodes()
	go s.periodicallyCheckVersion()
	return s
}

func (s *session) end() {
	close(s.done)
	s.listener.Close()
	delta := myNetMeta.Update(s.id, netmeta.NodeMeta{myAddr, true})
	Broadcast(createNetMetaUpdateMsg(delta))
	// TODO disconnect all connected nodes
	s.wg.Wait()
}

func startNewSession() (string, error) {
	lAddr, err := net.ResolveTCPAddr("tcp", myAddr)
	if err != nil {
		return "", err
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		return "", err
	}
	//util.Debug("listening on ", lAddr.String())
	mySession = getNewSession(listener)
	return mySession.id, nil
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
			go handleNewConn(conn)
		}
	}
}

func handleIncomingNetMeta(meta netmeta.NetMeta) {

}

// TODO: not too sure how to organize yet
//       might want to have locks here or maybe in netmeta
func getLatestMeta() []byte {
	return myNetMeta.toJson()
}

func (node *Node) reconnect() {
	// TODO reconnect
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
	return nil
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
