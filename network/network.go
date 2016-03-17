// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up: 
// $ go get github.com/satori/go.uuid

package network

import(
	"../util"
	"net"
)

func listenForConn(addr *net.TCPAddr) {
	listener, err := net.ListenTCP("tcp", addr)
	util.CheckError(err)
	for {
		conn, _ := listener.Accept()
		handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
}

func Initialize(addr string) error {
	_, err := net.ResolveTCPAddr("tcp", addr)
	return err
}

// All the following functions assume an Initialize call has been made

func ConnectTo(remoteAddr string) error {
	// TODO
	return nil
}

func Broadcast() {

}