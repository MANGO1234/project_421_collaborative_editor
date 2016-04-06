package treedocmanager

import (
	. "../common"
	"../treedoc2"
)

type LogEntry struct {
	Operation treedoc2.Operation
	Id        SiteId
	Version   uint32
}

type OperationLog struct {
	Log []LogEntry
}

func NewLog() *OperationLog {
	return &OperationLog{make([]LogEntry, 0, 100)}
}

func (log *OperationLog) Write(id SiteId, version uint32, operation treedoc2.Operation) {
	log.Log = append(log.Log, LogEntry{
		Id:        id,
		Version:   version,
		Operation: operation,
	})
}
