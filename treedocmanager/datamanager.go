package treedocmanager

// this part of treedocmanager deals with declaration & management of internal data needed by treedocmanager

import (
	. "../common"
	"../treedoc2"
	"../version"
	"encoding/binary"
	"github.com/satori/go.uuid"
	"sync"
)

type OperationID [20]byte

type versionInfo struct {
	sync.RWMutex
	versionQueue *version.VectorQueue
}

type operationLog struct {
	sync.RWMutex
	operations map[OperationID]treedoc2.Operation
}

// following fields keep track of treedoc management info
var (
	myVersionInfo  versionInfo
	myOperationLog operationLog
)

// initialize internal fields
func InitializedFields() {
	myVersionInfo = newVersionInfo()
	myOperationLog = newOperationLog()
}

// given uuid and clock, return node id
func NewNodeId(uuid SiteId, clock uint32) treedoc2.NodeId {
	versionByte := uint32ToBytes(clock)
	var ID [20]byte
	copy(ID[:], uuid[:])
	copy(ID[16:], versionByte)
	return ID
}

// given a operation id, return its uuid and version number
func SeparateOperationID(id OperationID) ([16]byte, uint32) {
	var uuid [16]byte
	copy(uuid[:], id[:16])
	version := bytesToUint32(id[16:])
	return uuid, version
}

// return a new queue elem to be broadcasted
func NewQueueElem(id OperationID, operation treedoc2.Operation) version.QueueElem {
	uuid, versionNum := SeparateOperationID(id)
	versionVect := myVersionInfo.getVersionVector()
	queueElem := version.QueueElem{Vector: versionVect, Id: uuid, Version: versionNum, Operation: operation}
	return queueElem
}

// update management data after local operation is performed
func UpdateVersion(opID OperationID, operation treedoc2.Operation) {
	// add operations to op log
	myOperationLog.add(opID, operation)
	// increment local version info
	uuid, version := SeparateOperationID(opID)
	myVersionInfo.update(uuid, version)
}

func (opLog operationLog) add(opID OperationID, operation treedoc2.Operation) {
	opLog.Lock()
	opLog.operations[opID] = operation
	opLog.Unlock()
}

func (versioninfo versionInfo) getVersionVector() version.VersionVector {
	versioninfo.RLock()
	defer versioninfo.RUnlock()
	return versioninfo.versionQueue.Vector()
}

func (versioninfo versionInfo) update(id SiteId, version uint32) {
	versioninfo.Lock()
	defer versioninfo.Unlock()
	versioninfo.versionQueue.IncrementVector(id, version)
}

func (operationId OperationID) toNodeId() treedoc2.NodeId {
	var id treedoc2.NodeId
	copy(id[:], operationId[:])
	return id
}

// *********  helpers  ****** //
func uint32ToBytes(num uint32) []byte {
	newByte := make([]byte, 4)
	binary.BigEndian.PutUint32(newByte, num)
	return newByte
}

func bytesToUint32(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(bytes)
}

func uuidToBytes(id string) []byte {
	givenUUID, _ := uuid.FromString(id)
	bytesUUID := givenUUID.Bytes()
	return bytesUUID
}

func newVersionInfo() versionInfo {
	newQueue := version.NewQueue()
	return versionInfo{versionQueue: newQueue}
}

func newOperationLog() operationLog {
	return operationLog{operations: make(map[OperationID]treedoc2.Operation)}
}
