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

func (n *node) handleNodeQuit() {
	// TODO
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
