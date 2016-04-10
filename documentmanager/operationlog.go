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

func (log *OperationLog) GetMissingOperations(vector version.VersionVector) []RemoteOperation {
	result := make([]RemoteOperation, 0, 10)
	checkDone := make(map[SiteId]bool, len(vector))
	emptyVector := version.NewVector()

	for i := len(log.Log) - 1; i >= 0; i-- {
		currentLog := log.Log[i]
		givenVersion := vector.Get(currentLog.Id)
		if givenVersion < currentLog.Version {
			result = append(result, RemoteOperation{
				Vector:  emptyVector,
				Id:      currentLog.Id,
				Version: currentLog.Version,
				Op:      currentLog.Operation,
			})
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

	for i := len(result)/2 - 1; i >= 0; i-- {
		opp := len(result) - 1 - i
		result[i], result[opp] = result[opp], result[i]
	}
	return result
}
