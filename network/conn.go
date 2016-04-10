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
		logger:  n.logger,
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

func (n *node) resetConn() {
	n.conn = nil
	n.reader = nil
	n.writer = nil
}

func (n *node) close() {
	n.conn.Close()
}


func (n *node) writeMessageSlice(msg []byte) error {
	msgByte :=  n.logger.PrepareSend("send bytes", msg)
	err := n.writer.WriteMessageSlice(msgByte)
	return err
}

func (n *node) writeMessage(msg string) error {
	msgByte :=  n.logger.PrepareSend("send string", []byte(msg))
	//return n.writer.WriteMessage(msgByte)
	err := n.writer.WriteMessageSlice(msgByte)
	return err
}

func (n *node) readMessage() (string, error) {
	msg, err := n.reader.ReadMessageSlice()
	//msg, err := n.reader.ReadMessage()
	var incomingMessage []byte
	n.logger.UnpackReceive("receive string", msg, &incomingMessage)
	return string(msg), err
}

func (n *node) readMessageSlice() ([]byte, error) {
	msg, err := n.reader.ReadMessageSlice()
	var incomingMessage []byte
	n.logger.UnpackReceive("receive byte", msg, &incomingMessage)
	return msg, err
}

func (n *node) sendMessage(msg Message) error {
	return n.writeMessageSlice(msg.toJson())
}

func (n *node) receiveMessage() (Message, error) {
	msgJson, err := n.readMessageSlice()
	msg := new(Message)
	n.logger.UnpackReceive("receive message", msgJson, &msg)
	if err != nil {
		return Message{}, err
	}
	return newMessageFromJson(msgJson)
}
