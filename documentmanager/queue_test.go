package documentmanager

import (
	. "../version"
	"reflect"
	"runtime/debug"
	"testing"
)

func assertEqual(t *testing.T, exp, got interface{}) {
	if !reflect.DeepEqual(exp, got) {
		debug.PrintStack()
		t.Fatalf("Expecting '%v' got '%v'\n", exp, got)
	}
}

func NewTestVector(a, b, c uint32) VersionVector {
	v := NewVector()
	v.IncrementTo(A_ID, a)
	v.IncrementTo(B_ID, b)
	v.IncrementTo(C_ID, c)
	return v
}

func TestEnqueue1(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(QueueElem{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(5, 0, 0), Id: B_ID, Version: 1})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1})
	assertEqual(t, 4, len(result))
	assertEqual(t, 1, q.Size())
	assertEqual(t, uint32(4), q.Vector().Get(A_ID))
	assertEqual(t, uint32(0), q.Vector().Get(B_ID))
	assertEqual(t, uint32(0), q.Vector().Get(C_ID))

	assertEqual(t, A_ID, result[0].Id)
	assertEqual(t, A_ID, result[1].Id)
	assertEqual(t, A_ID, result[2].Id)
	assertEqual(t, A_ID, result[3].Id)
	assertEqual(t, uint32(1), result[0].Version)
	assertEqual(t, uint32(2), result[1].Version)
	assertEqual(t, uint32(3), result[2].Version)
	assertEqual(t, uint32(4), result[3].Version)
}

func TestEnqueue2(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(QueueElem{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 1})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1})
	assertEqual(t, 5, len(result))
	assertEqual(t, 0, q.Size())
	assertEqual(t, uint32(4), q.Vector().Get(A_ID))
	assertEqual(t, uint32(1), q.Vector().Get(B_ID))
	assertEqual(t, uint32(0), q.Vector().Get(C_ID))

	assertEqual(t, A_ID, result[0].Id)
	assertEqual(t, A_ID, result[1].Id)
	assertEqual(t, A_ID, result[2].Id)
	assertEqual(t, A_ID, result[3].Id)
	assertEqual(t, B_ID, result[4].Id)
	assertEqual(t, uint32(1), result[0].Version)
	assertEqual(t, uint32(2), result[1].Version)
	assertEqual(t, uint32(3), result[2].Version)
	assertEqual(t, uint32(4), result[3].Version)
	assertEqual(t, uint32(1), result[4].Version)
}

func TestEnqueue3(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(QueueElem{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1})
	assertEqual(t, 1, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2})
	assertEqual(t, 1, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3})
	assertEqual(t, 1, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4})
	assertEqual(t, 1, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 1})
	assertEqual(t, 0, q.Size())
	assertEqual(t, uint32(4), q.Vector().Get(A_ID))
	assertEqual(t, uint32(1), q.Vector().Get(B_ID))
	assertEqual(t, uint32(0), q.Vector().Get(C_ID))
}

func TestEnqueue4(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(QueueElem{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(4, 2, 1), Id: A_ID, Version: 5})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(0, 2, 0), Id: C_ID, Version: 1})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 1})
	assertEqual(t, 0, len(result))
	result = q.Enqueue(QueueElem{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1})
	assertEqual(t, 5, len(result))
	assertEqual(t, 2, q.Size())
	assertEqual(t, uint32(4), q.Vector().Get(A_ID))
	assertEqual(t, uint32(1), q.Vector().Get(B_ID))
	assertEqual(t, uint32(0), q.Vector().Get(C_ID))

	result = q.Enqueue(QueueElem{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 2})
	assertEqual(t, 3, len(result))
	assertEqual(t, 0, q.Size())
	assertEqual(t, uint32(5), q.Vector().Get(A_ID))
	assertEqual(t, uint32(2), q.Vector().Get(B_ID))
	assertEqual(t, uint32(1), q.Vector().Get(C_ID))

	assertEqual(t, B_ID, result[0].Id)
	assertEqual(t, C_ID, result[1].Id)
	assertEqual(t, A_ID, result[2].Id)
	assertEqual(t, uint32(2), result[0].Version)
	assertEqual(t, uint32(1), result[1].Version)
	assertEqual(t, uint32(5), result[2].Version)
}

func TestEnqueueDuplicateOperations(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(QueueElem{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2})
	assertEqual(t, 0, len(result))
	assertEqual(t, 1, q.Size())
	result = q.Enqueue(QueueElem{Vector: NewTestVector(1, 0, 1), Id: A_ID, Version: 2})
	assertEqual(t, 0, len(result))
	assertEqual(t, 2, q.Size())
	result = q.Enqueue(QueueElem{Vector: NewTestVector(1, 0, 2), Id: A_ID, Version: 2})
	assertEqual(t, 0, len(result))
	assertEqual(t, 3, q.Size())
	result = q.Enqueue(QueueElem{Vector: NewTestVector(0, 0, 0), Id: C_ID, Version: 1})
	assertEqual(t, 1, len(result))
	assertEqual(t, 3, q.Size())
	result = q.Enqueue(QueueElem{Vector: NewTestVector(0, 0, 1), Id: C_ID, Version: 2})
	assertEqual(t, 1, len(result))
	assertEqual(t, 3, q.Size())
	result = q.Enqueue(QueueElem{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1})
	assertEqual(t, 2, len(result))
	assertEqual(t, 0, q.Size())
	assertEqual(t, uint32(2), q.Vector().Get(A_ID))
	assertEqual(t, uint32(0), q.Vector().Get(B_ID))
	assertEqual(t, uint32(2), q.Vector().Get(C_ID))
}
