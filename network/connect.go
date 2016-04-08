package network

import "encoding/json"

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

func (s *session) establishConnection(n *node) error {
	msg := s.getLatestVersionCheckMsg()
	rawMsg, _ := json.Marshal(msg)
	err := n.writeMessageSlice(rawMsg)
	if err != nil {
		return err
	}
	addToConnectedPool(n)
	go s.foreverRead(n)
	return nil
}

func (s *session) foreverRead(n *node) {
	s.Add(1)
	defer s.Done()
	for {
		rawMsg, err := n.readMessageSlice()
		if s.ended() {
			return
		}
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
