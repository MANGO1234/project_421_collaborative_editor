// This is a primitive try to handle network connections according to
// the proposal. If it works well, we can replace editor.go with this.
//
// Usage: go run node.go [ip:port]
// [ip:port] specify the ip and port for listening to incoming connections.

package main

import (
	"./network"
	// "./treedoc"
	"./treedocmanager"
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

	nodeId, err := network.Initialize(os.Args[1])
	treedocmanager.CreateTreedoc(nodeId)
	// TODO: link cursor position and treedoc posID

	util.CheckError(err)

	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		switch command {
		case "connect":
			fmt.Print("connect > ")
			scanner.Scan()
			remoteAddr := scanner.Text()
			id, err := network.ConnectTo(remoteAddr)
			fmt.Println("my id:", id)
			util.PrintError(err) // TODO deal with err
		case "disconnect":
			// TODO: disconnect to the network
			network.Disconnect()
		case "reconnect":
			network.Reconnect()
		case "insert":
			fmt.Print("insert > ")
			scanner.Scan()
			/*
				pos := scanner.Text()
				fmt.Print("insert " + pos + " > ")
				scanner.Scan()
			*/
			// txtInsert := scanner.Text()
			// if len(txtInsert) <= 1 {
			// 	treedoc.Insert(mydoc, currentPos, txtInsert[0])
			// } else {
			// 	tempDoc := treedoc.GenerateDoc(currentPos, txtInsert)
			// 	treedoc.Merge(mydoc, tempDoc)
			// }

			// fmt.Println(treedoc.DocToString(mydoc))

			// TODO: parse and insert
			//err = network.BroadCastInsert()
			// new line must be escaped on the client side
		case "delete":
			// fmt.Print("delete > ")
			// scanner.Scan()
			// pos := scanner.Text()
			// fmt.Print("delete " + pos + " > ")

			// treedoc.Delete(mydoc, currentPos)
			// fmt.Println(treedoc.DocToString(mydoc))

			//length := scanner.Text()
			// TODO: delete length characters at pos
			//err = network.BroadCastDelete()
		case "exportDoc":
			fmt.Print("exportDoc > ")
			scanner.Scan()
			//path := scanner.Text()
			// TODO: export the current doc into path
		case "printDoc":
			// TODO: print current doc
			//fmt.Println(treedoc.DocToString(mydoc))
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
