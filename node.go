// This is a primitive try to handle network connections according to
// the proposal. If it works well, we can replace editor.go with this.
//
// Usage: go run node.go [ip:port]
// [ip:port] specify the ip and port for listening to incoming connections.

package main

import (
	"./network"
	"./treedoc"
	"./util"
	"bufio"
	"fmt"
	"os"
)

func main() {
	// parse args.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [ip:port]\n", os.Args[0])
		os.Exit(1)
	}

	err := network.Initialize(os.Args[1])
	util.CheckError(err)

	// TODO initialize treedoc
	docRoot := treedoc.NewDisambiguatorNode()
	treedoc.Height(docRoot)

	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		switch command {
		case "connect":
			fmt.Print("connect > ")
			scanner.Scan()
			remoteAddr := scanner.Text()
			err = network.ConnectTo(remoteAddr)
			// TODO deal with err
		case "disconnect":
			// TODO: disconnect to the network
			network.Disconnect()
		case "insert":
			fmt.Print("insert > ")
			scanner.Scan()
			pos := scanner.Text()
			fmt.Print("insert " + pos + " > ")
			scanner.Scan()
			//string := scanner.Text()
			// TODO: parse and insert
			// new line must be escaped on the client side
		case "delete":
			fmt.Print("delete > ")
			scanner.Scan()
			pos := scanner.Text()
			fmt.Print("delete " + pos + " > ")
			scanner.Scan()
			//length := scanner.Text()
			// TODO: delete length characters at pos
		case "exportDoc":
			fmt.Print("exportDoc > ")
			scanner.Scan()
			//path := scanner.Text()
			// TODO: export the current doc into path
		case "printDoc":
			// TODO: print current doc
		case "printNetMeta":
			fmt.Println(network.GetNetworkMetadata())
		case "help":
			// TODO: print a help menu
		case "quit":
			// TODO: disconnect and gracefully quit
			os.Exit(0)
		default:
			fmt.Println("invalid command; enter \"help\" for help")
		}
		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
