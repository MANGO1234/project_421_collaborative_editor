package main

import (
	"./util"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

type Node struct {
	Addr net.Addr
	Id   string
}

var connectedNodes = struct {
	sync.RWMutex
	Nodes []Node
}{}

type NodeMetadata struct {
	Id    string
	Count uint64
	Addr  *net.Addr
	Node  []Node
}

var networkMetadata = make(map[string]NodeMetadata)

func listenForConn(addr *net.TCPAddr) {
	listener, err := net.ListenTCP("tcp", addr)
	util.CheckError(err)
	for {
		conn, _ := listener.Accept()
		handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	connectedNodes.Lock()
	connectedNodes.Nodes = append(connectedNodes.Nodes, Node{conn.RemoteAddr(), conn.RemoteAddr().String()})
	fmt.Println(connectedNodes)
	connectedNodes.Unlock()
}

func connectToRemote(laddr *net.TCPAddr, raddr *net.TCPAddr) {
	conn, err := net.Dial("tcp", raddr.String())
	util.CheckError(err)
	metadata := NodeMetadata{Addr: laddr, Count: 0, Node: laddr}

	messageWriter := util.MessageWriter{bufio.NewWriter(conn)}
	message, _ := json.Marshal(metadata)
	err = messageWriter.WriteMessageSlice(message)
	util.CheckError(err)

	connectedNodes.Lock()
	connectedNodes.Nodes = append(connectedNodes.Nodes, Node{conn.RemoteAddr(), conn.RemoteAddr().String()})
	fmt.Println(connectedNodes)
	connectedNodes.Unlock()
	messageReader := util.MessageReader{bufio.NewReader(conn)}
	for {
		messageReader.ReadMessage()
	}
}

func main() {
	var localAddr *net.TCPAddr
	var remoteAddr *net.TCPAddr
	var err error
	if len(os.Args) == 2 {
		localAddr, err = net.ResolveTCPAddr("tcp", os.Args[1])
		util.CheckError(err)
	} else if len(os.Args) == 3 {
		localAddr, err = net.ResolveTCPAddr("tcp", os.Args[1])
		util.CheckError(err)
		remoteAddr, err = net.ResolveTCPAddr("tcp", os.Args[2])
		util.CheckError(err)
	} else {
		os.Exit(0)
	}

	if remoteAddr != nil {
		go connectToRemote(localAddr, remoteAddr)
	}
	listenForConn(localAddr)
}
