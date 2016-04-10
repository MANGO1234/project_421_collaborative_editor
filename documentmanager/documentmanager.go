package documentmanager

import (
	"../buffer"
	. "../common"
	"../treedoc2"
	"../version"
	"fmt"
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
	BroadcastRemote func(RemoteOperation)
}

func NewDocumentModel(id SiteId, width int, updateGUI func()) *DocumentModel {
	return &DocumentModel{
		OwnerId:   id,
		Treedoc:   treedoc2.NewDocument(),
		Buffer:    buffer.StringToBuffer("", width),
		Queue:     NewQueue(),
		Log:       NewLog(),
		UpdateGUI: updateGUI,
	}
}

func (model *DocumentModel) LocalInsert(atom byte) {
	model.Lock()
	defer model.Unlock()
	pos := model.Buffer.GetPosition()
	model.Buffer.InsertAtCurrent(atom)
	id := treedoc2.NewNodeId(model.OwnerId, model.NodeIdClock)
	operation := treedoc2.InsertPos(model.Treedoc, id, pos, atom)
	if operation.Type == treedoc2.INSERT_NEW || operation.Type == treedoc2.INSERT_ROOT {
		model.NodeIdClock++
	}
	model.OpVersion++
	vector := model.Log.Vector.Copy()
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.AssertEqual()
	//model.Debug()
	if model.BroadcastRemote != nil {
		go model.BroadcastRemote(RemoteOperation{Vector: vector, Id: model.OwnerId, Version: model.OpVersion, Op: operation})
	}
}

func (model *DocumentModel) LocalBackspace() {
	model.Lock()
	defer model.Unlock()
	pos := model.Buffer.GetPosition() - 1
	if pos < 0 {
		return
	}
	model.Buffer.BackspaceAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	model.OpVersion++
	vector := model.Log.Vector.Copy()
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.AssertEqual()
	//model.Debug()
	if model.BroadcastRemote != nil {
		go model.BroadcastRemote(RemoteOperation{Vector: vector, Id: model.OwnerId, Version: model.OpVersion, Op: operation})
	}
}

func (model *DocumentModel) LocalDelete() {
	model.Lock()
	defer model.Unlock()
	pos := model.Buffer.GetPosition()
	if pos >= model.Buffer.GetSize() {
		return
	}
	model.Buffer.DeleteAtCurrent()
	operation := treedoc2.DeletePos(model.Treedoc, pos)
	model.OpVersion++
	vector := model.Log.Vector.Copy()
	model.Log.Write(model.OwnerId, model.OpVersion, operation)
	model.AssertEqual()
	//model.Debug()
	if model.BroadcastRemote != nil {
		go model.BroadcastRemote(RemoteOperation{Vector: vector, Id: model.OwnerId, Version: model.OpVersion, Op: operation})
	}
}

func (model *DocumentModel) ApplyRemoteOperation(op RemoteOperation) {
	model.Lock()
	defer model.Unlock()
	queueOps := model.Queue.Enqueue(op, model.Log.Vector.Copy())
	for _, queueOp := range queueOps {
		bufOp := model.Treedoc.ApplyOperation(queueOp.Op)
		model.Buffer.ApplyOperation(bufOp)
		model.Log.Write(queueOp.Id, queueOp.Version, queueOp.Op)
		model.AssertEqual()
	}
	//model.Debug()
	model.UpdateGUI()
}

func (model *DocumentModel) AssertEqual() {
	if model.Buffer.ToString() != treedoc2.DocToString(model.Treedoc) {
		panic("Not equal document!\n**************************\n" + model.Buffer.ToString() + "\n*******************\n" + treedoc2.DocToString(model.Treedoc))
	}
}

func (model *DocumentModel) Debug() {
	fmt.Println()
	fmt.Println()
	fmt.Println(model.OwnerId)
	fmt.Println(model.Log.Vector)
}

func (model *DocumentModel) SetBroadcastRemote(fn func(RemoteOperation)) {
	model.BroadcastRemote = fn
}

func (model *DocumentModel) RemoveBroadcastRemote() {
	model.BroadcastRemote = nil
}

func (model *DocumentModel) HandleVersionVector(vector version.VersionVector) {
	myVec := model.Log.Vector.Copy()
	compare := myVec.Compare(vector)
	if compare == version.GREATER_THAN || compare == version.CONFLICT {
		//ops := model.Log.GetMissingOperations(vector)
		//queueOps := model.Queue.GetMissingOperations(vector)
	}
}
