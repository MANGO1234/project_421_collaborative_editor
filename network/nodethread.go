package network

import (
	"net"
	"time"
)

func (s *session) initiateNewNode(n *node) {
	if n.state != nodeStateDisconnected {
		return
	}
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
		if s.ended() || n.state != nodeStateDisconnected || s.poke(n.id, n.addr) {
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
		s.manager.msgChan <- newNetMetaUpdateMsg(s.id, newQuitNetMeta(id, addr))
		return true
	}
}

func (s *session) connectThread(n *node) {
	// We establish actual connection when our addr is greater
	// than remote addr.
	n.interval = 0
	for {
		if s.ended() || n.state != nodeStateDisconnected || s.connect(n) {
			// connectThread stops when session ends
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
		if n.setState(nodeStateConnected) {
			go s.sendThread(getSendWrapperFromNode(n))
			go s.receiveThread(n)
		}
		return true
	} else {
		n.close()
		s.manager.msgChan <- newNetMetaUpdateMsg(s.id, newQuitNetMeta(n.id, n.addr))
		return true
	}
}

func (s *session) receiveThread(n *node) {
	for {
		if n.state != nodeStateConnected {
			n.close()
			return
		}
		msg, err := n.receiveMessage()
		if err != nil {
			n.close()
			n.setState(nodeStateDisconnected)
			if shouldConnect(s.manager.addr, n.addr) {
				go s.connectThread(n)
			}
			return
		}
		if s.ended() {
			// stops receiving msg once session ends
			return
		}
		s.manager.msgChan <- msg
	}
}

func (s *session) sendThread(sendWrapper *node) {
	for done := false; !done || len(sendWrapper.outChan) > 0; {
		select {
		case msg := <-sendWrapper.outChan:
			err := sendWrapper.sendMessage(msg)
			if err != nil {
				sendWrapper.close()
				sendWrapper.outChan <- msg
				return
			}
		case <-s.done:
			done = true
		}
	}
	sendWrapper.sendMessage(newNetMetaUpdateMsg(s.id, newQuitNetMeta(s.id, s.manager.addr)))
	sendWrapper.close()
}
