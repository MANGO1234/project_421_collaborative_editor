package documentmanager

import (
	"testing"
	"../treedoc2"
	. "../common"

)

var A_ID = StringToSiteId("aaaaaaaaaaaaaaaa")
var B_ID = StringToSiteId("bbbbbbbbbbbbbbbb")
var C_ID = StringToSiteId("cccccccccccccccc")

func TestGetMissingOperations(t *testing.T){
	log := NewLog()
	log.Write(A_ID,1,treedoc2.Operation{Id:treedoc2.NewNodeId(A_ID,1)})
	log.Write(A_ID,2,treedoc2.Operation{Id:treedoc2.NewNodeId(A_ID,2)})
	log.Write(B_ID,1,treedoc2.Operation{Id:treedoc2.NewNodeId(B_ID,1)})
	log.Write(C_ID,1,treedoc2.Operation{Id:treedoc2.NewNodeId(C_ID,1)})
	log.Write(B_ID,2,treedoc2.Operation{Id:treedoc2.NewNodeId(B_ID,2)})
	log.Write(A_ID,3,treedoc2.Operation{Id:treedoc2.NewNodeId(A_ID,3)})

	assertEqual(t,6,len(log.Log))

	v := NewTestVector(2,1,0)
	result := log.GetMissingOperations(v)
	assertEqual(t, 3, len(result))
	assertEqual(t,treedoc2.NewNodeId(C_ID,1),result[0].Id)
	assertEqual(t,treedoc2.NewNodeId(B_ID,2),result[1].Id)
	assertEqual(t,treedoc2.NewNodeId(A_ID,3),result[2].Id)

	v1 := NewTestVector(1,3,0)
	result = log.GetMissingOperations(v1)
	assertEqual(t,3,len(result))
	assertEqual(t,treedoc2.NewNodeId(A_ID,2),result[0].Id)
	assertEqual(t,treedoc2.NewNodeId(C_ID,1),result[1].Id)
	assertEqual(t,treedoc2.NewNodeId(A_ID,3),result[2].Id)
}



