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
// associated with the the uuid we have as quitted.
// this scheme avoids deadlock and inifinite connection replacement where we
// maintain only one connection between any two nodes
//
// For example, when A and B connect to each other simutaneously,
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

func handleNetworkError(err error, conn net.Conn) {
	if err != nil {
		conn.Close()
	}
}

func (s *session) handleNewConn(conn net.Conn) {
	wrapper := newConnWrapper(conn)
	// distinguish purpose of this connection
	purpose, err := wrapper.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}
	switch purpose {
	case dialingTypePoke:
		idMatches, err := handleIdCheck(s.id, wrapper)
		if err != nil || !idMatches {
			conn.Close()
			return
		}
		fallthrough
	case dialingTypeClientPoke:
		// send latest netmeta to the connecting node
		latestMeta := getLatestMeta()
		err = wrapper.WriteMessageSlice(latestMeta)
		if err != nil {
			return
		}
		incomingMeta, err := retrieveNetMeta(wrapper)
		if err != nil {
			return
		}
		conn.Close()
		handleIncomingNetMeta(incomingMeta)
	case dialingTypeConnect:
		idMatches, err := handleIdCheck(s.id, wrapper)
		if err != nil || !idMatches {
			conn.Close()
			return
		}
		n, err := retrieveNode(wrapper)
		if err != nil {
			return
		}
		establishConnection(n, wrapper)
	case dialingTypeClientConnect:
		err = wrapper.WriteMessage(s.id)
		if err != nil {
			conn.Close()
			return
		}
		n, err := retrieveNode(wrapper)
		if err != nil {
			return
		}
		establishConnection(n, wrapper)
	default:
		// invalid purpose
		conn.Close()
	}
}

func retrieveNode(wrapper *ConnWrapper) (*node, error) {
	id, err := wrapper.ReadMessage()
	if err != nil {
		wrapper.Close()
		return nil, err
	}
	addr, err := wrapper.ReadMessage()
	if err != nil {
		wrapper.Close()
		return nil, err
	}
	return addOrRetrieveNode(id, addr), nil
}

func addOrRetrieveNode(id, addr string) *node {
	// TODO
	// TODO check if it's already existing
	return &node{}
}

func foreverRead(wrapper *ConnWrapper) {
	for {
		rawMsg, err := wrapper.ReadMessageSlice()
		var msg Message
		if err != nil {
			handleDisconnect()
		}
		err = json.Unmarshal(rawMsg, &msg)
		if err != nil {
			handleBadNode()
		}
		myBroadcastChan <- msg
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
func register(localAddr, remoteAddr string) error {
	if shouldPoke(localAddr, remoteAddr) {
		return clientPoke(remoteAddr)
	} else {
		return clientConnect(remoteAddr)
	}
}

func (n *node) tryPokeOrConnect(localAddr string) error {
	if shouldPoke(localAddr, n.addr) {
		return n.poke()
	} else {
		return n.connect()
	}
}

func clientPoke(remoteAddr string) error {
	wrapper, err := dial(dialingTypeClientPoke, remoteAddr)
	if err != nil {
		return err
	}
	defer wrapper.Close()
	incomingMeta, err := retrieveNetMeta(wrapper)
	if err != nil {
		return err
	}
	// at this point, we consider the client poke as successful
	// since we have enough info to be considered as part of the network
	latestMeta := getLatestMeta()
	wrapper.WriteMessageSlice(latestMeta)
	handleIncomingNetMeta(incomingMeta)
	return nil
}

func (n *node) poke() error {
	wrapper, err := dial(dialingTypePoke, n.addr)
	if err != nil {
		return err
	}
	defer wrapper.Close()
	matches, err := checkId(wrapper, n.id)
	if err != nil {
		return err
	}
	if matches {
		incomingMeta, err := retrieveNetMeta(wrapper)
		if err != nil {
			return err
		}
		latestMeta := getLatestMeta()
		err = wrapper.WriteMessageSlice(latestMeta)
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

func clientConnect(remoteAddr string) error {
	wrapper, err := dial(dialingTypeClientConnect, remoteAddr)
	if err != nil {
		return err
	}
	remoteId, err := wrapper.ReadMessage()
	if err != nil {
		return err
	}
	n := addOrRetrieveNode(remoteId, remoteAddr)
	err = sendInfoAboutSelf(mySession.id, myAddr, wrapper)
	if err != nil {
		return err
	}
	return establishConnection(n, wrapper)
}

func (n *node) connect() error {
	wrapper, err := dial(dialingTypeClientConnect, n.addr)
	if err != nil {
		return err
	}
	matches, err := checkId(wrapper, n.id)
	if err != nil {
		wrapper.Close()
		return err
	}
	if matches {
		err = sendInfoAboutSelf(mySession.id, myAddr, wrapper)
		if err != nil {
			return err
		}
		return establishConnection(n, wrapper)
	} else {
		wrapper.Close()
		n.handleNodeQuit()
		return nil
	}
}

func sendInfoAboutSelf(id, addr string, wrapper *ConnWrapper) error {
	err := wrapper.WriteMessage(id)
	if err != nil {
		wrapper.Close()
		return err
	}
	err = wrapper.WriteMessage(addr)
	if err != nil {
		wrapper.Close()
	}
	return err
}

func establishConnection(n *node, wrapper *ConnWrapper) error {
	msg := createVersionCheckMsg()
	rawMsg, _ := json.Marshal(msg)
	err := wrapper.WriteMessageSlice(rawMsg)
	if err != nil {
		wrapper.Close()
		return err
	}
	addToConnectedPool(n, wrapper)
	go foreverRead(wrapper)
	return nil
}

func addToConnectedPool(n *node, wrapper *ConnWrapper) {
	// TODO
}

func shouldPoke(localAddr, remoteAddr string) bool {
	return localAddr > remoteAddr
}

func dial(dialType, remoteAddr string) (*ConnWrapper, error) {
	// connect to remote node
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}
	wrapper := newConnWrapper(conn)
	// indicate intention of this dial
	err = wrapper.WriteMessage(dialType)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return wrapper, nil
}

func checkId(wrapper *ConnWrapper, expectedId string) (success bool, err error) {
	err = wrapper.WriteMessage(expectedId)
	if err != nil {
		return
	}
	match, err := wrapper.ReadMessage()
	if err != nil {
		return
	}
	return match == "true", nil
}

func handleIdCheck(localId string, wrapper *ConnWrapper) (success bool, err error) {
	expectedId, err := wrapper.ReadMessage()
	if err != nil {
		return
	}
	if localId == expectedId {
		err = wrapper.WriteMessage("true")
		if err != nil {
			return
		}
		success = true
	} else {
		wrapper.WriteMessage("false")
		success = false
	}
	return
}

func retrieveNetMeta(wrapper *ConnWrapper) (NetMeta, error) {
	msg, err := wrapper.ReadMessageSlice()
	if err != nil {
		return nil, err
	}
	var incomingMeta NetMeta
	err = json.Unmarshal(msg, &incomingMeta)
	if err != nil {
		return incomingMeta, err
	}
	return incomingMeta, err
}

func handleIncomingNetMeta(meta NetMeta) {
	// TODO
}

// TODO: not too sure how to organize yet
//       might want to have locks here or maybe in netmeta
func getLatestMeta() []byte {
	return myNetMeta.ToJson()
}
