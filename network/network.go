// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

// TODO: debug mode to log any errors

package network

import (
	"../util"
	"./netmeta"
	"errors"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

type Node struct {
	addr string
	conn *ConnWrapper
}

type Message struct {
	Type    string
	Visited map[string]bool
	Msg     []byte
}

type ConnWrapper struct {
	conn   net.Conn
	reader *util.MessageReader
	writer *util.MessageWriter
}

type session struct {
	sync.WaitGroup
	id       string
	listener *net.TCPListener
	done     chan struct{}
}

// constants
// how often to check version vector and network metadata
const versionCheckInterval = 30

// how often to reconnect to disconnected nodes
const reconnectInterval = 30

// message types
const msgTypeVersionCheck = "versioncheck"
const msgTypeSync = "sync"
const msgTypeNetMetaUpdate = "netmeta"
const msgTypeTreedocOp = "treedocOp"

// states to keep track of
var myAddr string
var mySession *session
var myMsgChan chan Message
var myBroadcastChan chan Message
var myNetMeta netmeta.NetMeta

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
	go s.listenForNewConn()
	go s.periodicallyReconnectDisconnectedNodes()
	go s.periodicallyCheckVersion()
	return s
}

func (s *session) end() {
	close(s.done)
	s.listener.Close()
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

func handleNewConn(conn net.Conn) {
	// TODO
}

func (s *session) periodicallyReconnectDisconnectedNodes() {
	s.Add(1)
	defer s.Done()
	for {
		if s.ended() {
			return
		}
		reconnectDisconnectedNodes()
		time.Sleep(time.Second * reconnectInterval)
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
		checkVersion()
		time.Sleep(time.Second * versionCheckInterval)
	}
}

func checkVersion() {
	// TODO
}

// initialize local network listener
func Initialize(addr string) (string, error) {
	myAddr = addr
	myBroadcastChan = make(chan Message, 15)
	myMsgChan = make(chan Message)
	go serveBroadcastRequests(myBroadcastChan)
	go serveIncomingMessages(myMsgChan)
	return startNewSession(addr)
}

func serveIncomingMessages(in <-chan Message) {
	for msg := range in {
		//TODO
	}
}

func serveBroadcastRequests(in <-chan Message) {
	for msg := range in {
		//TODO
	}
}

// Disconnect from the network voluntarily
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

// All the following functions assume an Initialize call has been made
func ConnectTo(remoteAddr string) error {
	if mySession == nil {
		startNewSession() //TODO
	}
	// TODO
}

func Broadcast(msg Message) {
	go func() {
		myBroadcastChan <- msg
	}()
}

func GetNetworkMetadata() string {
	// TODO
}
