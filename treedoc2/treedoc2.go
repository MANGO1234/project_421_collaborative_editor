package treedoc2

import (
	"bytes"
	"fmt"
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

const INSERT_NEW byte = 0
const INSERT_ROOT byte = 1
const INSERT byte = 2
const DELETE byte = 3

type Operation struct {
	Type     byte
	ParentId NodeId
	ParentN  uint16
	Id       NodeId
	N        uint16
	Atom     byte
}

func NewDocument() *Document {
	return &Document{Doc: make([]*DocNode, 0, 4), Nodes: make(map[NodeId]*DocNode)}
}

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

func ApplyOperation(doc *Document, operation Operation) {
	if operation.Type == INSERT_NEW {
		InsertNew(doc, operation)
	} else if operation.Type == INSERT {
		Insert(doc, operation)
	} else if operation.Type == DELETE {
		Delete(doc, operation)
	} else if operation.Type == INSERT_ROOT {
		InsertRoot(doc, operation)
	}
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

func InsertNew(doc *Document, operation Operation) {
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
}

func InsertRoot(doc *Document, operation Operation) {
	newNode := &DocNode{
		Parent: nil,
		NodeId: operation.Id,
	}
	doc.Nodes[operation.Id] = newNode
	newNode.Atoms = extendAtoms(newNode.Atoms, operation.N)
	newNode.Atoms[operation.N] = Atom{Atom: operation.Atom, State: ALIVE, Size: 1}
	doc.Doc = insertNode(doc.Doc, newNode)
	updateSize(doc, newNode, 1)
}

func Delete(doc *Document, operation Operation) {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State != ALIVE {
		panic("Atom is not alive \"" + DocToString(doc) + "\" ")
	}
	node.Atoms[operation.N] = Atom{State: DEAD, Atom: atom.Atom, Left: atom.Left, Size: atom.Size - 1}
	updateSize(doc, node, -1)
}

func Insert(doc *Document, operation Operation) {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State != UNINITIALIZED {
		panic("Atom is not uninitialized \"" + DocToString(doc) + "\" ")
	}
	node.Atoms[operation.N] = Atom{State: ALIVE, Atom: operation.Atom, Left: atom.Left, Size: atom.Size + 1}
	updateSize(doc, node, 1)
}

func InsertPos(doc *Document, userId NodeId, pos int, ch byte) {
	if doc.Size == 0 {

	}
}

func DeletePos(doc *Document, pos int) {
	acc := 0
	var currentNode *DocNode
	var currentN int

	for _, node := range doc.Doc {
		currentNode = node
		if acc+node.Size > pos {
			break
		}
		acc += node.Size
	}

	for i, atom := range currentNode.Atoms {
		currentN = i
		if acc+atom.Size > pos {
			break
		}
		acc += atom.Size
	}

	for !(acc+currentNode.Atoms[currentN].Size-1 == pos && currentNode.Atoms[currentN].State == ALIVE) {
		for _, node := range currentNode.Atoms[currentN].Left {
			currentNode = node
			if acc+node.Size > pos {
				break
			}
			acc += node.Size
		}

		for i, atom := range currentNode.Atoms {
			currentN = i
			if acc+atom.Size > pos {
				break
			}
			acc += atom.Size
		}
	}

	atom := currentNode.Atoms[currentN]
	currentNode.Atoms[currentN] = Atom{State: DEAD, Atom: atom.Atom, Left: atom.Left, Size: atom.Size - 1}
	updateSize(doc, currentNode, -1)
}

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
