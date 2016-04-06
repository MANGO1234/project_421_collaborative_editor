package treedocmanager

import (
	"../buffer"
	"../treedoc2"
)

type DocumentModel struct {
	nodeIdClock uint32
	uuid        string
	Treedoc     *treedoc2.Document
	Buffer      *buffer.Buffer
}

func NewDocumentModel(uuid string, width int) *DocumentModel {
	newDoc := treedoc2.NewDocument()
	myDoc = newDoc
	InitializedFields(uuid)
	return &DocumentModel{
		uuid:    uuid,
		Treedoc: treedoc2.NewDocument(),
		Buffer:  buffer.StringToBuffer("a", width),
	}
}

func (model *DocumentModel) LocalInsert(atom byte) {
	pos := model.Buffer.GetPosition()
	model.Buffer.InsertAtCurrent(atom)
	id := NewOperationID(myUUID, model.nodeIdClock)
	operation := treedoc2.InsertPos(myDoc, id.toNodeId(), pos, atom)
	if operation.Type == treedoc2.INSERT_NEW || operation.Type == treedoc2.INSERT_ROOT {
		model.nodeIdClock++
	}
	myOpVersion++
	BroadcastOperation(myOpVersion, operation)
}

func (model *DocumentModel) LocalBackspace() {
	pos := model.Buffer.GetPosition() - 1
	if pos < 0 {
		return
	}
	model.Buffer.BackspaceAtCurrent()
	operation := treedoc2.DeletePos(myDoc, pos)
	myOpVersion++
	BroadcastOperation(myOpVersion, operation)
}

func (model *DocumentModel) LocalDelete() {
	pos := model.Buffer.GetPosition()
	if pos >= model.Buffer.GetSize() {
		return
	}
	model.Buffer.DeleteAtCurrent()
	operation := treedoc2.DeletePos(myDoc, pos)
	myOpVersion++
	BroadcastOperation(myOpVersion, operation)
}

func BroadcastOperation(operationVersion uint32, operation treedoc2.Operation) {

}
