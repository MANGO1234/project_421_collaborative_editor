package main

import (
	. "./common"
	"./gui"
	"./treedoc2"
	"./version"
	"time"
)

func main() {
	go func() {
		ID := StringToSiteId("aaaaaaaaaaaaaaaa")
		A_0 := treedoc2.StringToNodeId("aaaaaaaaaaaaaaaa0000")
		time.Sleep(time.Second)
		gui.Model().RemoteOperation(version.NewVector(), ID, 1, treedoc2.Operation{
			Id:   A_0,
			N:    0,
			Type: treedoc2.INSERT_ROOT,
			Atom: 'a',
		})
		gui.Model().UpdateGUI()
		time.Sleep(time.Second)
		gui.Model().RemoteOperation(version.NewVector(), ID, 2, treedoc2.Operation{
			Id:   A_0,
			N:    1,
			Type: treedoc2.INSERT,
			Atom: 'b',
		})
		gui.Model().UpdateGUI()
		time.Sleep(time.Second)
		gui.Model().RemoteOperation(version.NewVector(), ID, 3, treedoc2.Operation{
			Id:   A_0,
			N:    2,
			Type: treedoc2.INSERT,
			Atom: 'c',
		})
		gui.Model().RemoteOperation(version.NewVector(), ID, 4, treedoc2.Operation{
			Id:   A_0,
			N:    3,
			Type: treedoc2.INSERT,
			Atom: 'd',
		})
		gui.Model().UpdateGUI()
		time.Sleep(time.Second)
		gui.Model().RemoteOperation(version.NewVector(), ID, 5, treedoc2.Operation{
			Id:   A_0,
			N:    0,
			Type: treedoc2.DELETE,
		})
		gui.Model().UpdateGUI()
	}()
	gui.InitEditor(StringToSiteId("aaaaaaaaaaaaaaaa"))
	gui.CloseEditor()
}
