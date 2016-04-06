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
		Buffer:  buffer.StringToBuffer("", width),
	}
}

func (model *DocumentModel) LocalInsert(atom byte) {
	pos := model.Buffer.GetPosition()
	model.Buffer.InsertAtCurrent(atom)
	id := NewOperationID(myUUID, model.nodeIdClock)
	operation := treedoc2.InsertPos(model.Treedoc, id.toNodeId(), pos, atom)
	if operation.Type == treedoc2.INSERT_NEW || operation.Type == treedoc2.INSERT_ROOT {
		model.nodeIdClock++
	}
	myOpVersion++
	model.AssertEqual()
	BroadcastOperation(myOpVersion, operation)
}

func (model *DocumentModel) LocalBackspace() {
	pos := model.Buffer.GetPosition() - 1
	if pos < 0 {
		return
	}
	model.Buffer.BackspaceAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	myOpVersion++
	model.AssertEqual()
	BroadcastOperation(myOpVersion, operation)
}

func (model *DocumentModel) LocalDelete() {
	pos := model.Buffer.GetPosition()
	if pos >= model.Buffer.GetSize() {
		return
	}
	model.Buffer.DeleteAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	myOpVersion++
	model.AssertEqual()
	BroadcastOperation(myOpVersion, operation)
}

func (model *DocumentModel) AssertEqual() {
	if model.Buffer.ToString() != treedoc2.DocToString(model.Treedoc) {
		panic("Not equal document!\n**************************\n" + model.Buffer.ToString() + "\n*******************\n" + treedoc2.DocToString(model.Treedoc))
	}
}

func BroadcastOperation(operationVersion uint32, operation treedoc2.Operation) {

}
