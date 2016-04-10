package network

import "net"

func (s *session) handleNewConn(conn net.Conn) {
	n := newConnWrapper(conn)
	// distinguish purpose of this connection
	purpose, err := n.readMessage()
	if err != nil {
		conn.Close()
		return
	}
	switch purpose {
	case dialingTypeRegister:
		s.handleRegister(n)
	case dialingTypePoke:
		s.handlePoke(n)
	case dialingTypeConnect:
		s.handleConnect(n)
	default:
		// invalid purpose; ignore
	}
}

func (s *session) handleRegister(connWrapper *node) {
	defer connWrapper.close()
	latestNetMeta := s.manager.nodePool.getLatestNetMetaJson()
	err := connWrapper.writeMessageSlice(latestNetMeta)
	if err != nil {
		return
	}
	incomingNetMeta, err := connWrapper.readMessageSlice()
	if err != nil {
		return
	}
	incoming, err := newNetMetaFromJson(incomingNetMeta)
	if err != nil {
		return
	}
	s.manager.msgChan <- newNetMetaUpdateMsg(s.id, incoming)
}

func (s *session) handlePoke(connWrapper *node) {
	defer connWrapper.close()
	expectedId, err := connWrapper.readMessage()
	if err != nil {
		return
	}
	if s.id != expectedId {
		connWrapper.writeMessage("false")
		return
	}
	err = connWrapper.writeMessage("true")
	if err != nil {
		return
	}
	id, err := connWrapper.readMessage()
	if err != nil {
		return
	}
	addr, err := connWrapper.readMessage()
	if err != nil {
		return
	}
	s.manager.msgChan <- newNetMetaUpdateMsg(s.id, newJoinNetMeta(id, addr))
	// TODO: not sure if the following is necessary when using tcp
	// but it gives more guarantees
	connWrapper.writeMessage("done") // best we can do
}

func (s *session) handleConnect(connWrapper *node) {
	expectedId, err := connWrapper.readMessage()
	defer func(err error, connWrapper *node) {
		if err != nil {
			connWrapper.close()
		}
	}(err, connWrapper)
	if err != nil {
		return
	}
	if s.id != expectedId {
		connWrapper.writeMessage("false")
		return
	}
	err = connWrapper.writeMessage("true")
	if err != nil {
		return
	}
	id, err := connWrapper.readMessage()
	if err != nil {
		return
	}
	addr, err := connWrapper.readMessage()
	if err != nil {
		return
	}
	err = connWrapper.sendMessage(s.getLatestVersionCheckMsg())
	if err != nil {
		return
	}
	n := s.manager.nodePool.addOrGetNodeFromPool(id, NodeMeta{addr, false})
	n.conn = connWrapper.conn
	n.reader = connWrapper.reader
	n.writer = connWrapper.writer
	n.state = nodeStateConnected
	go s.sendThread(getSendWrapperFromNode(n))
	go s.receiveThread(n)
}
