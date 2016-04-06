package treedoc2

import (
	"../buffer"
	"bytes"
	"fmt"
	"math"
	"strconv"
)

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

func (doc *Document) ApplyOperation(operation Operation) buffer.BufferOperation {
	if operation.Type == INSERT_NEW {
		return doc.InsertNew(operation)
	} else if operation.Type == INSERT {
		return doc.Insert(operation)
	} else if operation.Type == DELETE {
		return doc.Delete(operation)
	} else if operation.Type == INSERT_ROOT {
		return doc.InsertRoot(operation)
	}
	return buffer.BufferOperation{Type: buffer.NO_OPERATION}
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

// ***************************************************************************************
// ******************** Operations On Treedoc (Remote operations) ************************
// ***************************************************************************************

func (doc *Document) InsertNew(operation Operation) buffer.BufferOperation {
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
	return buffer.BufferOperation{Type: buffer.INSERT, Pos: pos, Atom: operation.Atom}
}

func (doc *Document) InsertRoot(operation Operation) buffer.BufferOperation {
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
	return buffer.BufferOperation{Type: buffer.INSERT, Pos: pos, Atom: operation.Atom}
}

func (doc *Document) Delete(operation Operation) buffer.BufferOperation {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State != ALIVE {
		panic("Atom is not alive \"" + DocToString(doc) + "\" ")
	}
	pos := calcPos(doc, node, int(operation.N))
	node.Atoms[operation.N] = Atom{State: DEAD, Atom: atom.Atom, Left: atom.Left, Size: atom.Size - 1}
	updateSize(doc, node, -1)
	return buffer.BufferOperation{Type: buffer.DELETE, Pos: pos}
}

func (doc *Document) Insert(operation Operation) buffer.BufferOperation {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtoms(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State != UNINITIALIZED {
		panic("Atom is not uninitialized \"" + DocToString(doc) + "\" ")
	}
	node.Atoms[operation.N] = Atom{State: ALIVE, Atom: operation.Atom, Left: atom.Left, Size: atom.Size + 1}
	updateSize(doc, node, 1)
	pos := calcPos(doc, node, int(operation.N))
	return buffer.BufferOperation{Type: buffer.INSERT, Pos: pos, Atom: operation.Atom}
}

// ***************************************************************************************
// ******************** Operations From buffer (Local operations) ************************
// ***************************************************************************************

func nextNonEmptyNode(i int, nodes []*DocNode) (int, *DocNode) {
	for i = i + 1; i < len(nodes); i++ {
		if nodes[i].Size != 0 {
			return i, nodes[i]
		}
	}
	return i, nil
}

func nextNonEmptyAtom(i int, atoms []Atom) int {
	for i = i + 1; i < len(atoms); i++ {
		if atoms[i].Size != 0 {
			return i
		}
	}
	return i
}

func immediateNextEmptyUninitializedNode(i int, atoms []Atom) int {
	for i = i + 1; i < len(atoms); i++ {
		if atoms[i].Size != 0 || atoms[i].State == ALIVE {
			return -1
		}
		if atoms[i].State == UNINITIALIZED {
			return i
		}
	}
	if i >= math.MaxInt16-1 {
		return -1
	}
	return i
}

func findNodePos(pos int, acc int, nodes []*DocNode) (int, *DocNode) {
	i, node := nextNonEmptyNode(-1, nodes)
	for i < len(nodes) {
		if acc+node.Size > pos {
			return acc, node
		}
		acc += node.Size
		i, node = nextNonEmptyNode(i, nodes)
	}
	return acc, nil
}

func findAtomPos(pos int, acc int, atoms []Atom) (int, int) {
	i := nextNonEmptyAtom(-1, atoms)
	for i < len(atoms) {
		if acc+atoms[i].Size > pos {
			return acc, i
		}
		acc += atoms[i].Size
		i = nextNonEmptyAtom(i, atoms)
	}
	return acc, i
}

// traverse the tree until it finds the ALIVE atom at pos
func posToIdForDel(nodes []*DocNode, pos int) (*DocNode, int) {
	var currentN int
	acc, currentNode := findNodePos(pos, 0, nodes)
	acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	for !(acc+currentNode.Atoms[currentN].Size-1 == pos && currentNode.Atoms[currentN].State == ALIVE) {
		acc, currentNode = findNodePos(pos, acc, currentNode.Atoms[currentN].Left)
		acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	}
	return currentNode, currentN
}

// traverse the tree until it finds a place to insert (either a new node or reuse an applicable node)
func insertPosNewHelper(doc *Document, node *DocNode, n int, nodeId NodeId, ch byte) Operation {
	for {
		if len(node.Atoms[n].Left) == 0 {
			op := Operation{Type: INSERT_NEW, Id: nodeId, N: 0, ParentId: node.NodeId, ParentN: uint16(n), Atom: ch}
			doc.InsertNew(op)
			return op
		}
		node = node.Atoms[n].Left[len(node.Atoms[n].Left)-1]
		n = len(node.Atoms)
		if node.Atoms[n-1].State == UNINITIALIZED {
			n = n - 1
		}
		node.Atoms = extendAtoms(node.Atoms, uint16(n))
	}
}

func InsertPos(doc *Document, nodeId NodeId, pos int, ch byte) Operation {
	if doc.Size == 0 {
		op := Operation{Type: INSERT_ROOT, Id: nodeId, N: 0, Atom: ch}
		doc.InsertRoot(op)
		return op
	}

	var currentN int
	acc, currentNode := findNodePos(pos, 0, doc.Doc)
	// check for end of the document
	if currentNode == nil {
		currentNode := doc.Doc[len(doc.Doc)-1]
		currentN := len(currentNode.Atoms)
		if currentNode.Atoms[currentN-1].State == UNINITIALIZED {
			currentN = currentN - 1
		}
		currentNode.Atoms = extendAtoms(currentNode.Atoms, uint16(currentN))
		return insertPosNewHelper(doc, currentNode, currentN, nodeId, ch)
	}
	acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	for {
		//		if acc+currentNode.Atoms[currentN].Size-1 == pos {
		//			if equalId(currentNode.NodeId, nodeId) && currentNode.Atoms[currentN].State == UNINITIALIZED {
		//				op := Operation{Type: INSERT, Id: currentNode.NodeId, N: uint16(currentN), Atom: ch}
		//				Insert(doc, op)
		//				return op
		//			}
		//		}
		if acc+currentNode.Atoms[currentN].Size-1 == pos && currentNode.Atoms[currentN].State == ALIVE {
			return insertPosNewHelper(doc, currentNode, currentN, nodeId, ch)
		}
		acc, currentNode = findNodePos(pos, acc, currentNode.Atoms[currentN].Left)
		acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	}
	return Operation{Type: NO_OPERATION}
}

func DeletePos(doc *Document, pos int) Operation {
	node, n := posToIdForDel(doc.Doc, pos)
	op := Operation{Type: DELETE, Id: node.NodeId, N: uint16(n)}
	doc.Delete(op)
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
