package network

import "net"

func (s *session) handleNewConn(conn net.Conn) {
	n := newNodeFromConn(conn)
	// distinguish purpose of this connection
	purpose, err := n.readMessage()
	if err != nil {
		conn.Close()
		return
	}
	switch purpose {
	case dialingTypePoke:
		idMatches, err := handleIdCheck(s.id, n)
		if err != nil || !idMatches {
			conn.Close()
			return
		}
		fallthrough
	case dialingTypeClientPoke:
		latestNetMeta := s.manager.nodePool.getLatestNetMetaJson()
		err = n.writeMessageSlice(latestNetMeta)
		if err != nil {
			conn.Close()
			return
		}
		incomingNetMeta, err := n.readMessageSlice()
		if err != nil {
			conn.Close()
			return
		}
		incoming, err := newNetMetaFromJson(incomingNetMeta)
		if err != nil {
			conn.Close()
			return
		}
		conn.Close()
		s.manager.msgChan <- newNetMetaUpdateMsg(s.id, incoming)
	case dialingTypeConnect:
		idMatches, err := handleIdCheck(s.id, n)
		if err != nil || !idMatches {
			conn.Close()
			return
		}
		err = n.retrieveIdAddr()
		if err != nil {
			conn.Close()
			return
		}
		s.establishConnection(n)
	case dialingTypeClientConnect:
		err = n.writeMessage(s.id)
		if err != nil {
			return
		}
		err = n.retrieveIdAddr()
		if err != nil {
			return
		}
		s.establishConnection(n)
	default:
		// invalid purpose; ignore
	}
}
