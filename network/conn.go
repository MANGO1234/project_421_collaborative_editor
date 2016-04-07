package network

import (
	"../util"
	"bufio"
	"net"
)

func (n *node) setConn(conn net.Conn) {
	n.conn = conn
	n.reader = &util.MessageReader{bufio.NewReader(conn)}
	n.writer = &util.MessageWriter{bufio.NewWriter(conn)}
}

func (n *node) resetConn() {
	n.conn = nil
	n.reader = nil
	n.writer = nil
}

func (n *node) close() error {
	return n.conn.Close()
}

// if err occurs, close conn, set state to disconnect
// return error so the caller can decide whether to do any other
// error handling
func (n *node) handleAndReturnError(err error) error {
	if err != nil {
		// TODO: test and see if this is working as expected
		//       ie. should not make a legitimate node leave
		if _, ok := err.(net.Error); !ok {
			// if it's not a network error, we assume the node to be an imposter
			// and force it to leave the system
			n.state = nodeStateLeft
		}
		n.close()
		n.resetConn()
		n.state = nodeStateDisconnected
	}
	return err
}

func (n *node) writeMessageSlice(msg []byte) error {
	err := n.writer.WriteMessageSlice(msg)
	return n.handleAndReturnError(err)
}

func (n *node) writeMessage(msg string) error {
	err := n.writer.WriteMessage(msg)
	return n.handleAndReturnError(err)
}

func (n *node) readMessage() (string, error) {
	msg, err := n.reader.ReadMessage()
	return msg, n.handleAndReturnError(err)
}

func (n *node) readMessageSlice() ([]byte, error) {
	msg, err := n.reader.ReadMessageSlice()
	return msg, n.handleAndReturnError(err)
}
