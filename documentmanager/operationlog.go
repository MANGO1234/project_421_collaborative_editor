package documentmanager

import (
	. "../common"
	"../treedoc2"
	"../version"
)

type LogEntry struct {
	Operation treedoc2.Operation
	Id        SiteId
	Version   uint32
}

type OperationLog struct {
	Log    []LogEntry
	Vector version.VersionVector
}

func NewLog() *OperationLog {
	return &OperationLog{make([]LogEntry, 0, 100), version.NewVector()}
}

func (log *OperationLog) Write(id SiteId, version uint32, operation treedoc2.Operation) {
	log.Vector.IncrementTo(id, version)
	log.Log = append(log.Log, LogEntry{
		Id:        id,
		Version:   version,
		Operation: operation,
	})
}
