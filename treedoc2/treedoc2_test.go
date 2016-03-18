package treedoc2

import (
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

func newId(id string) NodeId {
	var ID [20]byte
	copy(ID[:], id[:])
	return ID
}

var A_ID0 = newId("aaaaaaaaaaaaaaaa0000")
var A_ID1 = newId("aaaaaaaaaaaaaaaa0001")
var B_ID0 = newId("bbbbbbbbbbbbbbbb0000")
var B_ID1 = newId("bbbbbbbbbbbbbbbb0001")
var C_ID0 = newId("cccccccccccccccc0000")
var C_ID1 = newId("cccccccccccccccc0001")
var D_ID0 = newId("dddddddddddddddd0000")

func NewTestDoc() *Document {
	d := NewDocument()
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'e', ParentId: C_ID1, ParentN: 1, Id: D_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT, Atom: 'f', Id: C_ID0, N: 1})
	ApplyOperation(d, Operation{Type: INSERT, Atom: 'g', Id: C_ID1, N: 1})
	ApplyOperation(d, Operation{Type: INSERT, Atom: 'h', Id: C_ID1, N: 2})
	return d
}

// basic op
func TestInsert(t *testing.T) {
	d := NewDocument()
	assertEqual(t, -1, DocHeight(d))

	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	assertEqual(t, "ab", DocToString(d))
	assertEqual(t, 0, DocHeight(d))

	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'e', ParentId: C_ID1, ParentN: 1, Id: D_ID0, N: 0})
	assertEqual(t, "cadeb", DocToString(d))
	assertEqual(t, 2, DocHeight(d))

	ApplyOperation(d, Operation{Type: INSERT, Atom: 'f', Id: C_ID0, N: 1})
	ApplyOperation(d, Operation{Type: INSERT, Atom: 'g', Id: C_ID1, N: 1})
	ApplyOperation(d, Operation{Type: INSERT, Atom: 'h', Id: C_ID1, N: 2})
	assertEqual(t, "cfadeghb", DocToString(d))
	assertEqual(t, 2, DocHeight(d))

	d = NewDocument()
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'c', Id: C_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	assertEqual(t, "abc", DocToString(d))
}

func TestInsertDisambiguator(t *testing.T) {
	d := NewDocument()
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'c', Id: C_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	assertEqual(t, "abc", DocToString(d))

	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'c', ParentId: C_ID0, ParentN: 1, Id: C_ID1, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'a', ParentId: C_ID0, ParentN: 1, Id: A_ID1, N: 0})
	ApplyOperation(d, Operation{Type: INSERT_NEW, Atom: 'b', ParentId: C_ID0, ParentN: 1, Id: B_ID1, N: 0})
	assertEqual(t, "abcabc", DocToString(d))
}

func TestDelete(t *testing.T) {
	d := NewTestDoc()
	ApplyOperation(d, Operation{Type: DELETE, Id: C_ID1, N: 1})
	ApplyOperation(d, Operation{Type: DELETE, Id: C_ID0, N: 0})
	assertEqual(t, "fadehb", DocToString(d))
	x, y := DocStat(d)
	assertEqual(t, 6, x)
	assertEqual(t, 2, y)
}

//commutative
func TestInsertInsert(t *testing.T) {
	d1 := NewDocument()
	d2 := NewDocument()

	ApplyOperation(d1, Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	ApplyOperation(d1, Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	ApplyOperation(d2, Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	ApplyOperation(d2, Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	assertEqual(t, DocToString(d1), "ab")
	assertEqual(t, DocToString(d2), "ab")

	ApplyOperation(d1, Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	ApplyOperation(d1, Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	ApplyOperation(d2, Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	ApplyOperation(d2, Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	assertEqual(t, DocToString(d1), "cadb")
	assertEqual(t, DocToString(d2), "cadb")

	ApplyOperation(d1, Operation{Type: INSERT, Atom: 'e', Id: C_ID1, N: 1})
	ApplyOperation(d1, Operation{Type: INSERT, Atom: 'f', Id: C_ID1, N: 2})
	ApplyOperation(d2, Operation{Type: INSERT, Atom: 'f', Id: C_ID1, N: 2})
	ApplyOperation(d2, Operation{Type: INSERT, Atom: 'e', Id: C_ID1, N: 1})
	assertEqual(t, DocToString(d1), "cadefb")
	assertEqual(t, DocToString(d2), "cadefb")
}

func TestDeleteDelete(t *testing.T) {
	d := NewTestDoc()
	ApplyOperation(d, Operation{Type: DELETE, Id: C_ID1, N: 1})
	assertEqual(t, "cfadehb", DocToString(d))
	x, y := DocStat(d)
	assertEqual(t, 7, x)
	assertEqual(t, 1, y)

	ApplyOperation(d, Operation{Type: DELETE, Id: C_ID1, N: 1})
	assertEqual(t, "cfadehb", DocToString(d))
	x, y = DocStat(d)
	assertEqual(t, 7, x)
	assertEqual(t, 1, y)
}
