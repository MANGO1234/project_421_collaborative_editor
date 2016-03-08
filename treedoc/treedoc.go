package treedoc

import (
	"bytes"
	"fmt"
)

const LEFT = false
const RIGHT = true

const ALIVE = byte(0)
const DEAD = byte(1)
const UNINITIALIZED = byte(2)

type SiteId []byte

type Dir struct {
	Id  SiteId
	Dir bool
}

type PosId []Dir

type DisambiguatorNode struct {
	Nodes []*DocNode
}

type DocNode struct {
	SiteId SiteId
	Atom   byte
	State  byte
	Left   *DisambiguatorNode
	Right  *DisambiguatorNode
}

func NewDisambiguatorNode() *DisambiguatorNode {
	return &DisambiguatorNode{make([]*DocNode, 0, 4)}
}

func findNode(disambiguator *DisambiguatorNode, siteId SiteId) (int, *DocNode) {
	nodes := disambiguator.Nodes
	for i, node := range nodes {
		if bytes.Equal(node.SiteId, siteId) {
			return i, node
		}
	}
	return -1, nil
}

func insertNode(disambiguator *DisambiguatorNode, docNode *DocNode) {
	var i = 0
	for i = 0; i < len(disambiguator.Nodes); i++ {
		node := disambiguator.Nodes[i]
		result := bytes.Compare(node.SiteId, docNode.SiteId)
		if result > 0 {
			break
		} else if result == 0 {
			if node.State == UNINITIALIZED {
				node.Atom = docNode.Atom
				node.State = ALIVE
			}
			return
		}
	}

	nodes := append(disambiguator.Nodes, docNode)
	copy(nodes[i+1:], nodes[i:])
	nodes[i] = docNode
	disambiguator.Nodes = nodes
}

func deleteNode(disambiguator *DisambiguatorNode, siteId SiteId) {
	_, node := findNode(disambiguator, siteId)
	if node != nil {
		node.State = DEAD
	} else {
		insertNode(disambiguator, &DocNode{SiteId: siteId, State: DEAD})
	}
}

func navigateToDisambiguator(disambiguator *DisambiguatorNode, dirs []Dir) *DisambiguatorNode {
	for _, dir := range dirs {
		_, node := findNode(disambiguator, dir.Id)
		if node == nil {
			node = &DocNode{SiteId: dir.Id, State: UNINITIALIZED}
			insertNode(disambiguator, node)
		}

		if dir.Dir == LEFT {
			if node.Left == nil {
				node.Left = NewDisambiguatorNode()
			}
			disambiguator = node.Left
		} else {
			if node.Right == nil {
				node.Right = NewDisambiguatorNode()
			}
			disambiguator = node.Right
		}
	}
	return disambiguator
}

func Insert(disambiguator *DisambiguatorNode, posId PosId, atom byte) {
	length := len(posId)
	disambiguator = navigateToDisambiguator(disambiguator, posId[:length-1])
	siteId := posId[length-1].Id
	insertNode(disambiguator, &DocNode{Atom: atom, SiteId: siteId})
}

func Delete(disambiguator *DisambiguatorNode, posId PosId) {
	length := len(posId)
	disambiguator = navigateToDisambiguator(disambiguator, posId[:length-1])
	siteId := posId[length-1].Id
	deleteNode(disambiguator, siteId)
}

// disambiguator1 is mutated, disambiguator2 is not safe to use afterward
func Merge(disambiguator1 *DisambiguatorNode, disambiguator2 *DisambiguatorNode) {
	if disambiguator2 == nil {
		return
	}
	for _, node := range disambiguator2.Nodes {
		_, nodeFound := findNode(disambiguator1, node.SiteId)
		if nodeFound == nil {
			insertNode(disambiguator1, node)
		} else {
			if nodeFound.Left == nil && node.Left != nil {
				nodeFound.Left = NewDisambiguatorNode()
			}
			Merge(nodeFound.Left, node.Left)
			if nodeFound.Right == nil && node.Right != nil {
				nodeFound.Right = NewDisambiguatorNode()
			}
			Merge(nodeFound.Right, node.Right)
		}
	}
}

func GenerateDoc(posId PosId, str string) *DisambiguatorNode {
	root := NewDisambiguatorNode()
	length := len(posId)
	disambiguator := navigateToDisambiguator(root, posId[:length-1])
	siteId := posId[length-1].Id
	node := generateDocHelper(siteId, str, 0, len(str))
	if node != nil {
		disambiguator.Nodes = append(disambiguator.Nodes, node)
	}
	return root
}

func generateDocHelper(siteId SiteId, str string, a, b int) *DocNode {
	if a >= b {
		return nil
	}
	c := (a + b) / 2
	root := &DocNode{Atom: str[c], SiteId: siteId, State: ALIVE}
	leftNode := generateDocHelper(siteId, str, a, c)
	if leftNode != nil {
		root.Left = NewDisambiguatorNode()
		root.Left.Nodes = append(root.Left.Nodes, leftNode)
	}
	rightNode := generateDocHelper(siteId, str, c+1, b)
	if rightNode != nil {
		root.Right = NewDisambiguatorNode()
		root.Right.Nodes = append(root.Right.Nodes, rightNode)
	}
	return root
}

// debugging, printing etc.
func DebugDoc(disambiguator *DisambiguatorNode) {
	if disambiguator == nil {
		return
	}
	for _, node := range disambiguator.Nodes {
		DebugDoc(node.Left)
		if node.State == ALIVE {
			fmt.Print("  ")
			fmt.Printf("%q", node.Atom)
			fmt.Print(" ")
			fmt.Printf("%s\n", node.SiteId)
		} else if node.State == DEAD {
			fmt.Print("X ")
			fmt.Printf("%q", node.Atom)
			fmt.Print(" ")
			fmt.Printf("%s\n", node.SiteId)
		} else if node.State == UNINITIALIZED {
			fmt.Print("      ")
			fmt.Printf("%s\n", node.SiteId)
		}
		DebugDoc(node.Right)
	}
}

func DocToNodes(disambiguator *DisambiguatorNode) []*DocNode {
	return DocToNodesHelper(disambiguator, make([]*DocNode, 0, 10))
}

func DocToNodesHelper(disambiguator *DisambiguatorNode, slice []*DocNode) []*DocNode {
	if disambiguator == nil {
		return slice
	}
	for _, node := range disambiguator.Nodes {
		slice = DocToNodesHelper(node.Left, slice)
		slice = append(slice, node)
		slice = DocToNodesHelper(node.Right, slice)
	}
	return slice
}

func DocToBuffer(disambiguator *DisambiguatorNode) *bytes.Buffer {
	var buf bytes.Buffer
	return DocToBufferHelper(disambiguator, &buf)
}

func DocToBufferHelper(disambiguator *DisambiguatorNode, buf *bytes.Buffer) *bytes.Buffer {
	if disambiguator == nil {
		return buf
	}
	for _, node := range disambiguator.Nodes {
		buf = DocToBufferHelper(node.Left, buf)
		if node.State == ALIVE {
			buf.WriteByte(node.Atom)
		}
		buf = DocToBufferHelper(node.Right, buf)
	}
	return buf
}

func DocToString(disambiguator *DisambiguatorNode) string {
	return DocToBuffer(disambiguator).String()
}

func Height(disambiguator *DisambiguatorNode) int {
	if disambiguator == nil {
		return 0
	}
	max := 1
	for _, node := range disambiguator.Nodes {
		k := 1 + Height(node.Left)
		if k > max {
			max = k
		}
		k = 1 + Height(node.Right)
		if k > max {
			max = k
		}
	}
	return max
}
