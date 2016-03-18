package treedoc2

import (
	"bytes"
	"fmt"
	"math"
)

type NodeId [20]byte

type Document struct {
	Doc   []*DocNode
	Nodes map[NodeId]*DocNode
}

type DocNode struct {
	Parent *DocNode
	NodeId NodeId
	Atoms  []Atom
	Right  []*DocNode
}

const UNINITIATED byte = 0
const DEAD byte = 1
const ALIVE byte = 2

type Atom struct {
	State byte
	Atom  byte
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
	return &Document{make([]*DocNode, 0, 4), make(map[NodeId]*DocNode)}
}

func insertNode(disambiguator []*DocNode, node *DocNode) []*DocNode {
	if disambiguator == nil {
		a := make([]*DocNode, 1, 4)
		a[0] = node
		return a
	}

	var i = 0
	for i = len(disambiguator); i >= 1; i-- {
		docNode := disambiguator[i-1]
		if bytes.Compare(node.NodeId[:], docNode.NodeId[:]) > 0 {
			break
		}
	}

	disambiguator = append(disambiguator, node)
	copy(disambiguator[i:], disambiguator[i+1:])
	disambiguator[i] = node
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

func InsertNew(doc *Document, operation Operation) {
	parent := doc.Nodes[operation.ParentId]
	newNode := &DocNode{
		Parent: parent,
		NodeId: operation.Id,
	}
	newNode.Atoms = insertAtom(newNode.Atoms, Atom{Atom: operation.Atom, State: ALIVE}, operation.N)
	doc.Nodes[operation.Id] = newNode

	if operation.ParentN == math.MaxUint16 {
		parent.Right = insertNode(parent.Right, newNode)
	} else {
		parent.Atoms = extendAtoms(parent.Atoms, operation.ParentN)
		atom := parent.Atoms[operation.ParentN]
		parent.Atoms[operation.ParentN] = Atom{
			State: atom.State,
			Atom:  atom.Atom,
			Left:  insertNode(atom.Left, newNode),
		}
	}
}

func InsertRoot(doc *Document, operation Operation) {
	newNode := &DocNode{
		Parent: nil,
		NodeId: operation.Id,
	}
	doc.Nodes[operation.Id] = newNode
	newNode.Atoms = extendAtoms(newNode.Atoms, operation.N)
	newNode.Atoms[operation.N] = Atom{Atom: operation.Atom, State: ALIVE}
	doc.Doc = insertNode(doc.Doc, newNode)
}

func Delete(doc *Document, operation Operation) {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	node.Atoms[operation.N] = Atom{State: DEAD, Atom: atom.Atom, Left: atom.Left}
}

func Insert(doc *Document, operation Operation) {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	node.Atoms[operation.N] = Atom{State: ALIVE, Atom: operation.Atom, Left: atom.Left}
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
		buf = docToBufferHelper(node.Right, buf)
	}
	return buf
}

func DebugDoc(doc *Document) {
	DebugDocHelper(doc.Doc, "")
}

func DebugDocHelper(disambiguator []*DocNode, indent string) {
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			DebugDocHelper(atom.Left, indent+"  ")
			if atom.State == ALIVE {
				fmt.Print(indent)
				fmt.Print("  ")
				fmt.Printf("%q", atom.Atom)
				fmt.Print(" ")
				fmt.Printf("%s\n", node.NodeId)
			} else if atom.State == DEAD {
				fmt.Print(indent)
				fmt.Print(" x")
				fmt.Printf("%q", atom.Atom)
				fmt.Print(" ")
				fmt.Printf("%s\n", node.NodeId)
			}
		}
		DebugDocHelper(node.Right, indent+"  ")
	}
}
