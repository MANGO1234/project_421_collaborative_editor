package network

import (
	"net"
)

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

func (s *session) tryPokeOrConnect(n *node) error {
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
	incomingMeta, err := n.readMessageSlice()
	if err != nil {
		return err
	}
	// at this point, we consider the client poke as successful
	// since we have enough info to be considered as part of the network
	latestMeta := nm.nodePool.getLatestNetMetaJson()
	n.writeMessageSlice(latestMeta)
	nm.handleIncomingNetMeta(incomingMeta)
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
		incomingMeta, err := n.readMessageSlice()
		if err != nil {
			return err
		}
		latestMeta := s.manager.nodePool.getLatestNetMetaJson()
		err = n.writeMessageSlice(latestMeta)
		s.manager.handleIncomingNetMeta(incomingMeta)
		if err != nil {
			return err
		}
		n.close()
		return nil
	} else {
		s.manager.nodePool.forceNodeQuit(n)
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
