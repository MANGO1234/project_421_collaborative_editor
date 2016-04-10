package documentmanager

import "encoding/json"
import "../treedoc2"
import "../version"
import (
	. "../common"
	"github.com/nsf/termbox-go"
)

type RemoteOperation struct {
	Vector  version.VersionVector
	Id      SiteId
	Version uint32
	Op      treedoc2.Operation
}

type RemoteOperationJson struct {
	Vector  version.VersionVectorJson
	Id      SiteId
	Version uint32
	Op      treedoc2.Operation
}

func RemoteOperationToSlice(op RemoteOperation) []byte {
	slice, err := json.Marshal(RemoteOperationJson{
		Vector:  op.Vector.ToJsonable(),
		Id:      op.Id,
		Version: op.Version,
		Op:      op.Op,
	})
	if err != nil {
		termbox.Close()
		panic(err.Error())
	}
	return slice
}

func RemoteOperationFromSlice(slice []byte) RemoteOperation {
	var opJson RemoteOperationJson
	json.Unmarshal(slice, &opJson)
	return RemoteOperation{
		Vector:  version.FromVersionVectorJson(opJson.Vector),
		Id:      opJson.Id,
		Version: opJson.Version,
		Op:      opJson.Op,
	}
}
