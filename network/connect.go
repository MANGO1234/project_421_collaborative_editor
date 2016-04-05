package network

import (
	"encoding/json"
	"errors"
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

func register(remoteAddr string) error {
	return registerOrPokeHelper("", remoteAddr)
}

// this method doesn't try to establish a persistent connection
// it's goal is to register into the remote network and communicate
// the netmeta state between the two networks
// The actual persisting connection is to be established later
//

func handleNewConn(conn net.Conn) {
	defer conn.Close()
	wrapper := newConnWrapper(conn)
	// distinguish purpose of this connection
	purpose, err := wrapper.reader.ReadMessage()
	if err != nil {
		return
	}
	switch purpose {
	case dialingTypePoke:
		expectedId, err := wrapper.reader.ReadMessage()
		if err != nil {
			return
		}
		if expectedId == mySession.id {
			err = wrapper.writer.WriteMessageSlice("true")
			if err != nil {
				return
			}
		} else {
			wrapper.writer.WriteMessageSlice("false")
			return
		}
		fallthrough
	case dialingTypeRegister:
		// send latest netmeta to the connecting node
		latestMeta := getLatestMeta()
		err = wrapper.writer.WriteMessageSlice(latestMeta)
		if err != nil {
			return
		}
		// retrieve the latest netmeta from the connecting node
		msg, err := wrapper.reader.ReadMessageSlice()
		if err != nil {
			return
		}
		conn.Close()
		var incomingMeta netmeta.NetMeta
		err = json.Unmarshal(registrationJson, &incomingMeta)
		if err != nil {
			return
		}
		// establish persistent connection and perform any necessary broadcast
		handleIncomingNetMeta(incomingMeta)
	case dialingTypeEstablishConnection:
		// TODO: actually accept persistent
	default:
		// invalid purpose
		return
	}
}

func registerOrPokeHelper(id, remoteAddr string) error {
	// connect to remote node
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	wrapper := newConnWrapper(conn)
	// indicate intention of this dial
	var intention string
	if id == "" {
		intention = dialingTypeRegister
	} else {
		intention = dialingTypePoke
	}
	err = wrapper.writer.WriteMessage(intention)
	if err != nil {
		return err
	}
	// when poking, we need to make sure the other node has the expected id
	// if it doesn't we should treat the node associated the id as quitted
	if intention == dialingTypePoke {
		err = wrapper.writer.WriteMessage(id)
		if err != nil {
			return err
		}
		match, err := wrapper.reader.ReadMessage()
		if err != nil {
			return err
		}
		if match == "false" {
			// TODO: mark node as deleted
			return nil
		}
	}
	// retrieve latest netmeta from remote node
	msg, err := wrapper.reader.ReadMessageSlice()
	if err != nil {
		return err
	}
	var incomingMeta netmeta.NetMeta
	err = json.Unmarshal(registrationJson, &incomingMeta)
	if err != nil {
		return err
	}
	// send latest netmeta to the remote node
	latestMeta := getLatestMeta()
	wrapper.writer.WriteMessageSlice(latestMeta)
	// The write is this node's best attempt, with the netmeta received, this node
	// considers itself to be part of the network; Thus no error handling
	conn.Close()
	// establish persistent connection and perform any necessary broadcast
	handleIncomingNetMeta(incomingMeta)
}
