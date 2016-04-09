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

func (log *OperationLog) GetMissingOperations(vector version.VersionVector) []treedoc2.Operation {
	result := make([]treedoc2.Operation, 0)
	checkDone := make(map[SiteId]bool, len(vector))

	for i := len(log.Log) - 1; i >= 0; i-- {
		currentLog := log.Log[i]
		givenVersion := vector.Get(currentLog.Id)
		if givenVersion < currentLog.Version {
			result = append(result, currentLog.Operation)
			copy(result[1:], result[:])
			result[0] = currentLog.Operation
		} else {
			checkDone[currentLog.Id] = true
		}

		done := true

		for id, _ := range vector {
			if !checkDone[id] {
				done = false
			}
		}

		if done {
			break
		}
	}
	return result
}
