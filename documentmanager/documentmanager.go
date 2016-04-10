package documentmanager

import (
	"../buffer"
	. "../common"
	"../treedoc2"
	"../version"
	"sync"
)

type DocumentModel struct {
	sync.RWMutex
	OwnerId         SiteId
	OpVersion       uint32
	NodeIdClock     uint32
	Treedoc         *treedoc2.Document
	Buffer          *buffer.Buffer
	Log             *OperationLog
	Queue           *OperationQueue
	UpdateGUI       func()
	BroadcastRemote func(uint32, treedoc2.Operation)
}

func NewDocumentModel(id SiteId, width int, updateGUI func()) *DocumentModel {
	return &DocumentModel{
		OwnerId:         id,
		OpVersion:       1,
		Treedoc:         treedoc2.NewDocument(),
		Buffer:          buffer.StringToBuffer("", width),
		Queue:           NewQueue(),
		Log:             NewLog(),
		UpdateGUI:       updateGUI,
		BroadcastRemote: func(a uint32, b treedoc2.Operation) {},
	}
}

func (model *DocumentModel) LocalInsert(atom byte) {
	model.Lock()
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
	model.Unlock()
	go model.BroadcastRemote(model.OpVersion, operation)
}

func (model *DocumentModel) LocalBackspace() {
	model.Lock()
	pos := model.Buffer.GetPosition() - 1
	if pos < 0 {
		return
	}
	model.Buffer.BackspaceAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.OpVersion++
	model.AssertEqual()
	model.Unlock()
	go model.BroadcastRemote(model.OpVersion, operation)
}

func (model *DocumentModel) LocalDelete() {
	model.Lock()
	pos := model.Buffer.GetPosition()
	if pos >= model.Buffer.GetSize() {
		return
	}
	model.Buffer.DeleteAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.OpVersion++
	model.AssertEqual()
	model.Unlock()
	go model.BroadcastRemote(model.OpVersion, operation)
}

func (model *DocumentModel) RemoteOperation(vector version.VersionVector, id SiteId, opVersion uint32, operation treedoc2.Operation) {
	model.Lock()
	queueElems := model.Queue.Enqueue(QueueElem{
		Vector:    vector,
		Id:        id,
		Version:   opVersion,
		Operation: operation,
	})
	for _, elem := range queueElems {
		bufOp := model.Treedoc.ApplyOperation(elem.Operation)
		model.Buffer.ApplyOperation(bufOp)
		model.Log.Write(elem.Id, elem.Version, elem.Operation)
		model.AssertEqual()
	}
	model.Unlock()
	model.UpdateGUI()
}

func (model *DocumentModel) AssertEqual() {
	if model.Buffer.ToString() != treedoc2.DocToString(model.Treedoc) {
		panic("Not equal document!\n**************************\n" + model.Buffer.ToString() + "\n*******************\n" + treedoc2.DocToString(model.Treedoc))
	}
}

func (model *DocumentModel) SetBroadcastRemote(fn func(uint32, treedoc2.Operation)) {
	model.BroadcastRemote = fn
}

func (model *DocumentModel) RemoveBroadcastRemote() {
	model.BroadcastRemote = func(a uint32, b treedoc2.Operation) {}
}
