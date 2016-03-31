// This acts as the treedoc manager and manages the creation and modification of treedoc
// following methods are based on treedoc version 1.

import (
    "fmt"
    "../treedoc"
)

// following fields keep track of treedoc info
var (
	myDoc *treedoc.Disambiguator
    currentPos treedoc.PosId
)

// given the nodeId, create a new treedoc
func createTreedoc(String nodeId) error {
	newDoc := treedoc.GenerateDoc([]treedoc.Dir{
		treedoc.Dir{[]byte(nodeId), false}}, "")
	myDoc = newDoc
	return nil
}

func insertTo(){

}

func deleteFrom(){

}

func changePos(){

}

func getCurrentPos(){

}

