package treedocmanager

// This acts as the treedoc manager and manages the creation and modification of treedoc
// following methods are based on treedoc version 2.

import (
	"../treedoc2"
	"fmt"
)

// given the nodeId, create a new treedoc
func CreateTreedoc(uuid string) {
	newDoc := treedoc2.NewDocument()
	myDoc = newDoc
	InitializedFields(uuid)
}

// insert a single char at cursorPos
func Insert(atom byte, cursorPos int) {
	myOpVersion++
	id := NewOperationID(myUUID, myOpVersion)
	operation := treedoc2.InsertPos(myDoc, id.toNodeId(), cursorPos, atom)
	broadcastOperation(id, operation)
	UpdateVersion(id, operation)
}

// delete a single at cursorPos
func Delete(cursorPos int, length int) {
	myOpVersion++
	id := NewOperationID(myUUID, myOpVersion)
	operation := treedoc2.DeletePos(myDoc, cursorPos)
	broadcastOperation(id, operation)
	UpdateVersion(id, operation)
}

// broadcast local operations to other nodes
func broadcastOperation(id OperationID, operation treedoc2.Operation) {

}

// print the treedoc as string
func PrintTreedoc() {
	fmt.Println(treedoc2.DocToString(myDoc))
}
