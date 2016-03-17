package treedoc2

import (
	"bytes"
	"fmt"
	"math"
)

type SiteId struct {
	Id [16]byte
	N  uint16
}

type Document struct {
	Doc   []*DocNode
	Nodes map[SiteId]*DocNode
}

type DocNode struct {
	Parent *DocNode
	SiteId SiteId
	Atoms  []Atom
	Right  []*DocNode
}

const UNKNOWN byte = 0
const DEAD byte = 1
const ALIVE byte = 2

type Atom struct {
	State byte
	Atom  byte
	Left  []*DocNode
}

const INSERT_NEW byte = 0
const INSERT byte = 1
const DELETE byte = 2

type Operation struct {
	Type     byte
	ParentId SiteId
	Id       SiteId
	Atom     byte
}

func NewDocument() *Document {
	return &Document{make([]*DocNode, 0, 4), make(map[SiteId]*DocNode)}
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
		result := bytes.Compare(node.SiteId.Id[:], docNode.SiteId.Id[:])
		if result > 0 {
			break
		}
	}

	disambiguator = append(disambiguator, node)
	copy(disambiguator[i:], disambiguator[i+1:])
	disambiguator[i] = node
	return disambiguator
}

func extendAtoms(atoms []Atom, i uint16) []Atom {
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
		InsertN(doc, operation)
	} else if operation.Type == DELETE {
		DeleteN(doc, operation)
	}
}

func InsertNew(doc *Document, operation Operation) {
	parent := doc.Nodes[operation.ParentId]
	newNode := &DocNode{
		Parent: parent,
		SiteId: operation.Id,
		Atoms:  make([]Atom, 0, 4),
	}
	insertAtom(newNode.Atoms, Atom{Atom: operation.Atom, State: ALIVE}, operation.Id.N)

	if operation.ParentId.N == math.MaxUint16 {
		parent.Right = insertNode(parent.Right, newNode)
	} else {
		extendAtoms(parent.Atoms, operation.ParentId.N)
		atom := parent.Atoms[operation.ParentId.N]
		parent.Atoms[operation.ParentId.N] = Atom{
			State: atom.State,
			Atom:  atom.Atom,
			Left:  insertNode(atom.Left, newNode),
		}
	}
}

func DeleteN(doc *Document, operation Operation) {
	node := doc.Nodes[operation.Id]
	extendAtoms(node.Atoms, operation.Id.N)
	atom := node.Atoms[operation.Id.N]
	node.Atoms[operation.Id.N] = Atom{State: DEAD, Atom: atom.Atom, Left: atom.Left}
}

func InsertN(doc *Document, operation Operation) {
	node := doc.Nodes[operation.Id]
	extendAtoms(node.Atoms, operation.Id.N)
	atom := node.Atoms[operation.Id.N]
	node.Atoms[operation.Id.N] = Atom{State: ALIVE, Atom: atom.Atom, Left: atom.Left}
}

func DocToBuffer(doc *Document) *bytes.Buffer {
	var buf bytes.Buffer
	return docToBufferHelper(doc.Doc, &buf)
}

func DocToString(doc *Document) string {
	return DocToBuffer(doc).String()
}

func docToBufferHelper(disambiguator []*DocNode, buf *bytes.Buffer) *bytes.Buffer {
	if disambiguator == nil {
		return buf
	}
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
	if disambiguator == nil {
		return
	}
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			DebugDocHelper(atom.Left, indent+"  ")
			if atom.State == ALIVE {
				fmt.Print(indent)
				fmt.Print(" ")
				fmt.Print(atom.Atom)
			} else if atom.State == DEAD {
				fmt.Print(indent)
				fmt.Print("x")
				fmt.Print(atom.Atom)
			}
		}
		DebugDocHelper(node.Right, indent+"  ")
	}
}
