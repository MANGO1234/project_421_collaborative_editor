package documentmanager

import (
	"../buffer"
	. "../common"
	"../treedoc2"
	"../version"
	"encoding/json"
	"sync"
)

type DocumentModel struct {
	sync.RWMutex
	OwnerId     SiteId
	OpVersion   uint32
	NodeIdClock uint32
	Treedoc     *treedoc2.Document
	Buffer      *buffer.Buffer
	Log         *OperationLog
	Queue       *OperationQueue
	UpdateGUI   func()
}

type MissingOperation struct {
	missingOp   []treedoc2.Operation
	missingElem []QueueElem
}

func NewDocumentModel(id SiteId, width int, updateGUI func()) *DocumentModel {
	return &DocumentModel{
		OwnerId:   id,
		OpVersion: 1,
		Treedoc:   treedoc2.NewDocument(),
		Buffer:    buffer.StringToBuffer("", width),
		Queue:     NewQueue(),
		Log:       NewLog(),
		UpdateGUI: updateGUI,
	}
}

func (model *DocumentModel) LocalInsert(atom byte) {
	pos := model.Buffer.GetPosition()
	model.Buffer.InsertAtCurrent(atom)
	id := treedoc2.NewNodeId(model.OwnerId, model.NodeIdClock)
	operation := treedoc2.InsertPos(model.Treedoc, id, pos, atom)
	if operation.Type == treedoc2.INSERT_NEW || operation.Type == treedoc2.INSERT_ROOT {
		model.NodeIdClock++
	}
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.OpVersion++
	model.AssertEqual()
	BroadcastOperation(model.OwnerId, model.OpVersion, model.Queue.Vector(), operation)
}

func (model *DocumentModel) LocalBackspace() {
	pos := model.Buffer.GetPosition() - 1
	if pos < 0 {
		return
	}
	model.Buffer.BackspaceAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.OpVersion++
	model.AssertEqual()
	BroadcastOperation(model.OwnerId, model.OpVersion, model.Queue.Vector(), operation)
}

func (model *DocumentModel) LocalDelete() {
	pos := model.Buffer.GetPosition()
	if pos >= model.Buffer.GetSize() {
		return
	}
	model.Buffer.DeleteAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.OpVersion++
	model.AssertEqual()
	BroadcastOperation(model.OwnerId, model.OpVersion, model.Queue.Vector(), operation)
}

func (model *DocumentModel) RemoteOperation(newElem QueueElem) {
	queueElems := model.Queue.Enqueue(newElem)
	for _, elem := range queueElems {
		bufOp := model.Treedoc.ApplyOperation(elem.Operation)
		model.Buffer.ApplyOperation(bufOp)
		model.Log.Write(elem.Id, elem.Version, elem.Operation)
		model.AssertEqual()
	}
	model.UpdateGUI()
}

func (model *DocumentModel) AssertEqual() {
	if model.Buffer.ToString() != treedoc2.DocToString(model.Treedoc) {
		panic("Not equal document!\n**************************\n" + model.Buffer.ToString() + "\n*******************\n" + treedoc2.DocToString(model.Treedoc))
	}
}

func BroadcastOperation(id SiteId, opVersion uint32, vector version.VersionVector, operation treedoc2.Operation) {
	newQueueElem := QueueElem{
		Vector:    vector,
		Id:        id,
		Version:   opVersion,
		Operation: operation,
	}
	json.Marshal(newQueueElem)
	// TODO : call network to send queue element

}

func (model *DocumentModel) HandleVersionVector(vector version.VersionVector) {
	myVec := model.Queue.Vector()
	compare := myVec.Compare(vector)
	if compare == version.GREATER_THAN || compare == version.CONFLICT {
		ops := model.Log.GetMissingOperations(vector)
		queueElem := model.Queue.GetMissingQueueElem(vector)
		SendMissingOperations(MissingOperation{missingOp: ops, missingElem: queueElem})
	}
}

func SendMissingOperations(ops MissingOperation) {
	json.Marshal(ops)
	// TODO: call network to send missing operations
}
