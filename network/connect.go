package network

import (
	"encoding/json"
	"net"
)

// Note on the design:
//
// The node with smaller listening addr pokes    the node with a larger  one.
// The node with larger  listening addr connects the node with a smaller one.
// We distinguish between client-initiated vs. system-initiated dials because
// the client-initiated ones doesn't know the uuid of the other node. For the
// system-initiated dials, if the uuid of the node with the listening address
// changes, it means that the previous session ended either caused by a user
// initiated disconnect command or by stopping the node and restarting with
// same listening address. In this case, we need to consider the old session
// associated with the the uuid we have as left.
// this scheme avoids deadlock and infinite connection replacement where we
// maintain only one connection between any two nodes
//
// For example, when A and B connect to each other simultaneously,
// with a deterministic algorithm, it's possible that the two nodes
// performs symmetric actions, causing connections to be constantly
// replaced and failure to agree on one single connection between
// the nodes. If we try to use locks or channels to forbid this
// unnecessary reconnection loop, it's easy to get deadlock. Having
// two connections between the nodes solves this problem in a way
// but handling two connections when we only need to handle one is
// costly and leads to other problems associated with handling more
// connections

// the purpose of dialing to a node
const (
	// client-initiated poke to a remote node
	dialingTypeClientPoke = "clientpoke"
	// poke a known node so it has information to connect
	dialingTypePoke = "poke"
	// client-initialed connect to a remote node
	dialingTypeClientConnect = "clientconnect"
	// establish persistent connection between the nodes
	dialingTypeConnect = "connect"
)

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
		incomingMeta, err := n.retrieveNetMeta()
		if err != nil {
			return
		}
		conn.Close()
		handleIncomingNetMeta(incomingMeta)
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

func (n *node) retrieveIdAddr() error {
	id, err := n.readMessage()
	if err != nil {
		return err
	}
	n.id = id
	addr, err := n.readMessage()
	if err != nil {
		return err
	}
	n.addr = addr
	return nil
}

func (s *session) foreverRead(n *node) {
	s.Add(1)
	defer s.Done()
	for {
		rawMsg, err := n.readMessageSlice()
		if err != nil {
			// TODO reconnect the node; move to disconnected pool
			return
		}
		var msg Message
		err = json.Unmarshal(rawMsg, &msg)
		if err != nil {
			// TODO reconnect the node; move to disconnected pool
			//      quit the node
			return
		}
		s.manager.msgChan <- msg
	}
}

func handleBadNode() {
	// We can just ignore it; no harm done
	// TODO: if we have time, should force the node to quit
}

func handleDisconnect() {
	// TODO
}

// func poke(id, remoteAddr string) error {

// }

// client uses this to register itself to a remote network
func (nm *NetworkManager) register(remoteAddr string) error {
	if shouldPoke(nm.addr, remoteAddr) {
		return nm.clientPoke(remoteAddr)
	} else {
		return nm.clientConnect(remoteAddr)
	}
}

func (s *session) tryPokeOrConnect(n *node, localAddr string) error {
	if shouldPoke(s.manager.addr, n.addr) {
		return s.poke(n)
	} else {
		return s.connect(n)
	}
}

func (nm *NetworkManager) clientPoke(remoteAddr string) error {
	n := newNodeFromAddr(remoteAddr)
	err := n.dial(dialingTypeClientPoke)
	if err != nil {
		return err
	}
	incomingMeta, err := n.retrieveNetMeta()
	if err != nil {
		return err
	}
	// at this point, we consider the client poke as successful
	// since we have enough info to be considered as part of the network
	latestMeta := nm.nodePool.getLatestNetMetaJson()
	n.writeMessageSlice(latestMeta)
	handleIncomingNetMeta(incomingMeta)
	return nil
}

func (s *session) poke(n *node) error {
	err := n.dial(dialingTypePoke)
	if err != nil {
		return err
	}
	defer n.close()
	matches, err := n.checkId()
	if err != nil {
		return err
	}
	if matches {
		incomingMeta, err := n.retrieveNetMeta()
		if err != nil {
			return err
		}
		latestMeta := s.manager.nodePool.getLatestNetMetaJson()
		err = n.writeMessageSlice(latestMeta)
		if err != nil {
			handleIncomingNetMeta(incomingMeta)
			return err
		}
		handleIncomingNetMeta(incomingMeta)
		return nil
	} else {
		n.handleNodeQuit()
		return nil
	}
}

func (n *node) handleNodeQuit() {
	// TODO
}

func (nm *NetworkManager) clientConnect(remoteAddr string) error {
	n := newNodeFromAddr(remoteAddr)
	err := n.dial(dialingTypeClientConnect)
	if err != nil {
		return err
	}
	remoteId, err := n.readMessage()
	if err != nil {
		return err
	}
	n.id = remoteId
	err = sendInfoAboutSelf(nm.id, nm.addr, n)
	if err != nil {
		return err
	}
	return nm.session.establishConnection(n)
}

func (s *session) connect(n *node) error {
	err := n.dial(dialingTypeClientConnect)
	if err != nil {
		return err
	}
	matches, err := n.checkId()
	if err != nil {
		n.close()
		return err
	}
	if matches {
		err = sendInfoAboutSelf(s.id, s.manager.addr, n)
		if err != nil {
			return err
		}
		return s.establishConnection(n)
	} else {
		n.close()
		n.handleNodeQuit()
		return nil
	}
}

func sendInfoAboutSelf(id, addr string, n *node) error {
	err := n.writeMessage(id)
	if err != nil {
		return err
	}
	return n.writeMessage(addr)
}

func (s *session) establishConnection(n *node) error {
	msg := s.getLatestVersionCheckMsg()
	rawMsg, _ := json.Marshal(msg)
	err := n.writeMessageSlice(rawMsg)
	if err != nil {
		n.close()
		return err
	}
	addToConnectedPool(n)
	go s.foreverRead(n)
	return nil
}

func addToConnectedPool(n *node) {
	// TODO
}

func shouldPoke(localAddr, remoteAddr string) bool {
	return localAddr > remoteAddr
}

func (n *node) dial(dialType string) error {
	// connect to remote node
	conn, err := net.Dial("tcp", n.addr)
	if err != nil {
		return err
	}
	n.setConn(conn)
	// indicate intention of this dial
	return n.writeMessage(dialType)
}

func (n *node) checkId() (success bool, err error) {
	err = n.writeMessage(n.id)
	if err != nil {
		return
	}
	match, err := n.readMessage()
	if err != nil {
		return
	}
	return match == "true", nil
}

func handleIdCheck(localId string, n *node) (success bool, err error) {
	expectedId, err := n.readMessage()
	if err != nil {
		return
	}
	if localId == expectedId {
		err = n.writeMessage("true")
		if err != nil {
			return
		}
		success = true
	} else {
		n.writeMessage("false")
		success = false
	}
	return
}

func (n *node) retrieveNetMeta() (NetMeta, error) {
	msg, err := n.readMessageSlice()
	if err != nil {
		return nil, err
	}
	var incomingMeta NetMeta
	err = json.Unmarshal(msg, &incomingMeta)
	if err != nil {
		n.resetConn()
		n.state = nodeStateDisconnected
	}
	return incomingMeta, err
}

func handleIncomingNetMeta(meta NetMeta) {
	// TODO
}
