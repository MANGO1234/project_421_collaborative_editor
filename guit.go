package main

import (
	"./gui"
	"./network"
	"fmt"
	"os"
)

func main() {
	// parse args.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [ip:port]\n", os.Args[0])
		os.Exit(1)
	}

	networkManager, err := network.NewNetworkManager(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	gui.StartMainLoop(networkManager)
}
