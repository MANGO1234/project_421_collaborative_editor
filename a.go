package main

import (
	. "./treedoc2"
	"fmt"
)

var A_ID [20]byte
var B_ID [20]byte
var C_ID [20]byte

func main() {
	a := NewDocument()
	copy(A_ID[:], "aaaaaaaaaaaaaaaa0001"[:])
	copy(B_ID[:], "bbbbbbbbbbbbbbbb0001"[:])
	copy(C_ID[:], "cccccccccccccccc0001"[:])
	ApplyOperation(a, Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID, N: 0})
	ApplyOperation(a, Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID, N: 0})
	ApplyOperation(a, Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID, ParentN: 0, Id: C_ID, N: 0})
	ApplyOperation(a, Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID, ParentN: 1, Id: C_ID, N: 0})
	ApplyOperation(a, Operation{Type: INSERT, Atom: 'e', Id: C_ID, N: 1})
	ApplyOperation(a, Operation{Type: INSERT, Atom: 'f', Id: C_ID, N: 2})
	ApplyOperation(a, Operation{Type: DELETE, Id: C_ID, N: 2})
	fmt.Println(a)
	fmt.Println(a.Nodes[A_ID])
	DebugDoc(a)
	fmt.Println(DocToString(a))
}
