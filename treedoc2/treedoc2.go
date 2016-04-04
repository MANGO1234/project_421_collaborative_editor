package treedoc2

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
)

type NodeId [20]byte

type Document struct {
	Size  int
	Doc   []*DocNode
	Nodes map[NodeId]*DocNode
}

type DocNode struct {
	Parent  *DocNode
	ParentN uint16
	NodeId  NodeId
	Size    int
	Atoms   []Atom
}

const UNINITIALIZED byte = 0
const DEAD byte = 1
const ALIVE byte = 2

type Atom struct {
	State byte
	Atom  byte
	Size  int
	Left  []*DocNode
}

const NO_OPERATION byte = 0
const INSERT_NEW byte = 1
const INSERT_ROOT byte = 2
const INSERT byte = 3
const DELETE byte = 4

type Operation struct {
	Type     byte
	ParentId NodeId
	ParentN  uint16
	Id       NodeId
	N        uint16
	Atom     byte
}

type BufferOperation struct {
	Type byte
	Pos  int
	Atom byte
}

func NewDocument() *Document {
	return &Document{Doc: make([]*DocNode, 0, 4), Nodes: make(map[NodeId]*DocNode)}
}

// insert a node in sorted nodeId order into the slice
func insertNode(disambiguator []*DocNode, node *DocNode) []*DocNode {
	if disambiguator == nil {
		a := make([]*DocNode, 1, 4)
		a[0] = node
		return a
	}

	var a = 0
	for i := len(disambiguator); i >= 1; i-- {
		docNode := disambiguator[i-1]
		if bytes.Compare(node.NodeId[:], docNode.NodeId[:]) > 0 {
			a = i
			break
		}
	}

	disambiguator = append(disambiguator, node)
	copy(disambiguator[a+1:], disambiguator[a:])
	disambiguator[a] = node
	return disambiguator
}

// just ensure the atom slice has enough elements to prevent out of index error
func extendAtoms(atoms []Atom, i uint16) []Atom {
	if atoms == nil {
		return make([]Atom, i+1, i+5)
	}

	n := len(atoms)
	for int(i) >= n {
		atoms = append(atoms, Atom{})
		n++
	}
	return atoms
}

func insertAtom(atoms []Atom, atom Atom, i uint16) []Atom {
	atoms = extendAtoms(atoms, i)
	atoms[i] = atom
	return atoms
}

func ApplyOperation(doc *Document, operation Operation) BufferOperation {
	if operation.Type == INSERT_NEW {
		return InsertNew(doc, operation)
	} else if operation.Type == INSERT {
		return Insert(doc, operation)
	} else if operation.Type == DELETE {
		return Delete(doc, operation)
	} else if operation.Type == INSERT_ROOT {
		return InsertRoot(doc, operation)
	}
	return BufferOperation{Type: NO_OPERATION}
}

func updateSize(doc *Document, node *DocNode, delta int) {
	node.Size += delta
	if node.Parent == nil {
		doc.Size += delta
	} else {
		atom := node.Parent.Atoms[node.ParentN]
		node.Parent.Atoms[node.ParentN] = Atom{
			State: atom.State,
			Atom:  atom.Atom,
			Size:  atom.Size + delta,
			Left:  atom.Left,
		}
		updateSize(doc, node.Parent, delta)
	}
}

func calcPosHelper(doc *Document, node *DocNode, n int) int {
	acc := 0
	for i := 0; i < n; i++ {
		acc += node.Atoms[i].Size
	}
	if node.Parent == nil {
		for _, rootNode := range doc.Doc {
			if rootNode == node {
				break
			}
			acc += rootNode.Size
		}
		return acc
	}
	return acc + calcPosHelper(doc, node.Parent, int(node.ParentN))
}

// calculate the position of the atom given its parent node and its n
func calcPos(doc *Document, node *DocNode, n int) int {
	return calcPosHelper(doc, node, n) + node.Atoms[n].Size - 1
}

func InsertNew(doc *Document, operation Operation) BufferOperation {
	parent := doc.Nodes[operation.ParentId]
	newNode := &DocNode{
		NodeId:  operation.Id,
		Parent:  parent,
		ParentN: operation.ParentN,
	}
	newNode.Atoms = insertAtom(newNode.Atoms, Atom{
		State: ALIVE,
		Atom:  operation.Atom,
		Size:  1,
	}, operation.N)
	doc.Nodes[operation.Id] = newNode

	parent.Atoms = extendAtoms(parent.Atoms, operation.ParentN)
	atom := parent.Atoms[operation.ParentN]
	parent.Atoms[operation.ParentN] = Atom{
		State: atom.State,
		Atom:  atom.Atom,
		Size:  atom.Size,
		Left:  insertNode(atom.Left, newNode),
	}
	updateSize(doc, newNode, 1)
	pos := calcPos(doc, newNode, int(operation.N))
	return BufferOperation{Type: INSERT, Pos: pos, Atom: operation.Atom}
}

func InsertRoot(doc *Document, operation Operation) BufferOperation {
	newNode := &DocNode{
		Parent: nil,
		NodeId: operation.Id,
	}
	doc.Nodes[operation.Id] = newNode
	newNode.Atoms = extendAtoms(newNode.Atoms, operation.N)
	newNode.Atoms[operation.N] = Atom{Atom: operation.Atom, State: ALIVE, Size: 1}
	doc.Doc = insertNode(doc.Doc, newNode)
	updateSize(doc, newNode, 1)
	pos := calcPos(doc, newNode, int(operation.N))
	return BufferOperation{Type: INSERT, Pos: pos, Atom: operation.Atom}
}

func Delete(doc *Document, operation Operation) BufferOperation {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State != ALIVE {
		panic("Atom is not alive \"" + DocToString(doc) + "\" ")
	}
	pos := calcPos(doc, node, int(operation.N))
	node.Atoms[operation.N] = Atom{State: DEAD, Atom: atom.Atom, Left: atom.Left, Size: atom.Size - 1}
	updateSize(doc, node, -1)
	return BufferOperation{Type: DELETE, Pos: pos}
}

func Insert(doc *Document, operation Operation) BufferOperation {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State != UNINITIALIZED {
		panic("Atom is not uninitialized \"" + DocToString(doc) + "\" ")
	}
	node.Atoms[operation.N] = Atom{State: ALIVE, Atom: operation.Atom, Left: atom.Left, Size: atom.Size + 1}
	updateSize(doc, node, 1)
	pos := calcPos(doc, node, int(operation.N))
	return BufferOperation{Type: INSERT, Pos: pos, Atom: operation.Atom}
}

func findNodePos(pos int, acc int, nodes []*DocNode) (int, *DocNode) {
	for _, node := range nodes {
		if acc+node.Size > pos {
			return acc, node
		}
		acc += node.Size
	}
	return -1, nil
}

func findAtomPos(pos int, acc int, atoms []Atom) (int, int) {
	for i, atom := range atoms {
		if acc+atom.Size > pos {
			return acc, i
		}
		acc += atom.Size
	}
	return -1, 0
}

func insertNewPos(currentNode *DocNode, currentN int, acc, pos int, nodeId NodeId, ch byte) {

}

func InsertPos(doc *Document, nodeId NodeId, pos int, ch byte) Operation {
	if doc.Size == 0 {
		op := Operation{Type: INSERT_ROOT, Id: nodeId, N: 0, Atom: ch}
		InsertRoot(doc, op)
		return op
	}

	var currentNode *DocNode
	var currentN int
	acc := 0

	acc, currentNode = findNodePos(pos, acc, doc.Doc)
	acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	for {
		if acc+currentNode.Atoms[currentN].Size-1 == pos {
			if currentNode.Atoms[currentN].State == ALIVE {
				if bytes.Equal(currentNode.NodeId[0:16], nodeId[0:16]) && currentN < math.MaxUint16-1 {
					if (len(currentNode.Atoms) == currentN+1) ||
						(currentNode.Atoms[currentN+1].State == UNINITIALIZED && currentNode.Atoms[currentN+1].Size == 0) {
						op := Operation{Type: INSERT, Id: currentNode.NodeId, N: uint16(currentN + 1), Atom: ch}
						Insert(doc, op)
						return op
					}
				}
				op := Operation{Type: INSERT_NEW, Id: nodeId, N: 0, Atom: ch,
					ParentId: currentNode.NodeId, ParentN: uint16(currentN)}
				InsertNew(doc, op)
				return op
			} else if currentNode.Atoms[currentN].State == UNINITIALIZED && bytes.Equal(currentNode.NodeId[0:16], nodeId[0:16]) && currentN < math.MaxUint16 {
				op := Operation{Type: INSERT, Id: currentNode.NodeId, N: uint16(currentN), Atom: ch}
				Insert(doc, op)
				return op
			}
		}

		acc, currentNode = findNodePos(pos, acc, currentNode.Atoms[currentN].Left)
		acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	}
	return Operation{Type: NO_OPERATION}
}

func DeletePos(doc *Document, pos int) Operation {
	var currentNode *DocNode
	var currentN int
	acc := 0

	acc, currentNode = findNodePos(pos, acc, doc.Doc)
	acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	for !(acc+currentNode.Atoms[currentN].Size-1 == pos && currentNode.Atoms[currentN].State == ALIVE) {
		acc, currentNode = findNodePos(pos, acc, currentNode.Atoms[currentN].Left)
		acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	}

	op := Operation{Type: DELETE, Id: currentNode.NodeId, N: uint16(currentN)}
	Delete(doc, op)
	return op
}

// ***********************************************************
// *************** Miscellaneous *****************************
// ***********************************************************
func DocToBuffer(doc *Document) *bytes.Buffer {
	var buf bytes.Buffer
	return docToBufferHelper(doc.Doc, &buf)
}

func DocToString(doc *Document) string {
	return DocToBuffer(doc).String()
}

func docToBufferHelper(disambiguator []*DocNode, buf *bytes.Buffer) *bytes.Buffer {
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			buf = docToBufferHelper(atom.Left, buf)
			if atom.State == ALIVE {
				buf.WriteByte(atom.Atom)
			}
		}
	}
	return buf
}

func DebugDoc(doc *Document) {
	debugDocHelper(doc.Doc, "")
}

func debugDocHelper(disambiguator []*DocNode, indent string) {
	for _, node := range disambiguator {
		fmt.Print(indent + "  ")
		fmt.Printf("(%d)\n", node.Size)
		for _, atom := range node.Atoms {
			debugDocHelper(atom.Left, indent+"  ")
			if atom.State == ALIVE {
				fmt.Print(indent)
				fmt.Print("  ")
				fmt.Printf("%q", atom.Atom)
				fmt.Printf("(%d) ", atom.Size)
				fmt.Printf("%s\n", node.NodeId)
			} else if atom.State == DEAD {
				fmt.Print(indent)
				fmt.Print(" x")
				fmt.Printf("%q", atom.Atom)
				fmt.Printf("(%d) ", atom.Size)
				fmt.Print(" ")
				fmt.Printf("%s\n", node.NodeId)
			}
		}
	}
}

func DocHeight(doc *Document) int {
	return docHeightHelper(doc.Doc)
}

func docHeightHelper(disambiguator []*DocNode) int {
	max := -1
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			a := docHeightHelper(atom.Left) + 1
			if a > max {
				max = a
			}
		}
	}
	return max
}

func DocStat(doc *Document) (int, int) {
	return docStat(doc.Doc)
}

func docStat(disambiguator []*DocNode) (int, int) {
	alive := 0
	dead := 0
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			a, d := docStat(atom.Left)
			alive += a
			dead += d
			if atom.State == ALIVE {
				alive++
			} else if atom.State == DEAD {
				dead++
			}
		}
	}
	return alive, dead
}

func AtomToString(atom Atom) string {
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(int(atom.Atom)))
	buf.WriteString(" " + strconv.Itoa(int(atom.Size)))
	buf.WriteString(" " + strconv.Itoa(int(atom.State)))
	return buf.String()
}
