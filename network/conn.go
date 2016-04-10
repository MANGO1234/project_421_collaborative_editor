package network

import (
	"../util"
	"bufio"
	"net"
)

// These are methods on node which acts like a connection wrapper
// Structural wise, this is probably a misuse
// TODO fix the structure if possible

func getSendWrapperFromNode(n *node) *node {
	return &node{
		conn:    n.conn,
		reader:  n.reader,
		writer:  n.writer,
		outChan: n.outChan,
	}
}

func newConnWrapper(conn net.Conn) *node {
	var n node
	n.setConn(conn)
	return &n
}

func (n *node) setConn(conn net.Conn) {
	n.conn = conn
	n.reader = &util.MessageReader{bufio.NewReader(conn)}
	n.writer = &util.MessageWriter{bufio.NewWriter(conn)}
}

func (n *node) close() {
	n.conn.Close()
}

func (n *node) writeMessageSlice(msg []byte) error {
	return n.writer.WriteMessageSlice(msg)
}

func (n *node) writeMessage(msg string) error {
	return n.writer.WriteMessage(msg)
}

func (n *node) readMessage() (string, error) {
	return n.reader.ReadMessage()
}

func (n *node) readMessageSlice() ([]byte, error) {
	return n.reader.ReadMessageSlice()
}

func (n *node) sendMessage(msg Message) error {
	return n.writeMessageSlice(msg.toJson())
}

func (n *node) receiveMessage() (Message, error) {
	msgJson, err := n.readMessageSlice()
	if err != nil {
		return Message{}, err
	}
	return newMessageFromJson(msgJson)
}
