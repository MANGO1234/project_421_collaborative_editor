package treedocmanager

// This acts as the treedoc manager and manages the creation and modification of treedoc
// following methods are based on treedoc version 1.

import (
	"../treedoc"
)

// following fields keep track of treedoc info
var (
	myDoc      *treedoc.DisambiguatorNode
	currentPos treedoc.PosId
)

// given the nodeId, create a new treedoc
func CreateTreedoc(nodeId string) {
	newDoc := treedoc.GenerateDoc([]treedoc.Dir{
		treedoc.Dir{[]byte(nodeId), false}}, "")
	myDoc = newDoc
	currentPos = GetCurrentPos(0)
}

func InsertTo() {

}

func DeleteFrom() {

}

func ChangePos() {

}

func GetCurrentPos(cursorPos int) treedoc.PosId {
	return nil
}

func getTreedoc() *treedoc.DisambiguatorNode {
	return nil
}
