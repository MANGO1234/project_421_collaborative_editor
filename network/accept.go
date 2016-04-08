package network

import "net"

func (s *session) handleNewConn(conn net.Conn) {
	s.Add(1)
	defer s.Done()
	n := newNodeFromConn(conn)
	// distinguish purpose of this connection
	purpose, err := n.readMessage()
	if err != nil {
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
		// send latest netmeta to the connecting node
		latestMeta := s.manager.nodePool.getLatestNetMetaJson()
		err = n.writeMessageSlice(latestMeta)
		if err != nil {
			return
		}
		incomingMeta, err := n.readMessageSlice()
		if err != nil {
			return
		}
		incoming, err := newNetMetaFromJson(incomingMeta)
		if err != nil {
			return
		}
		conn.Close()
		s.manager.handleIncomingNetMeta(newNetMetaUpdateMsg(s.id, incoming))
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
