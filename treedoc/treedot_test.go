package treedoc

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

// basic op
func TestInsert(t *testing.T) {
	doc := NewDisambiguatorNode()
	Insert(doc, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	assertEqual(t, "a", DocToString(doc))
	Insert(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
	}, 'b')
	assertEqual(t, "ab", DocToString(doc))
	Insert(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
		Dir{[]byte("A"), false},
	}, 'c')
	assertEqual(t, "acb", DocToString(doc))
}

func TestDelete(t *testing.T) {
	doc := NewDisambiguatorNode()
	Insert(doc, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	Insert(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
	}, 'b')
	Insert(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
		Dir{[]byte("A"), false},
	}, 'c')
	assertEqual(t, "acb", DocToString(doc))

	Delete(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
		Dir{[]byte("A"), false},
	})
	assertEqual(t, "ab", DocToString(doc))
	Delete(doc, []Dir{
		Dir{[]byte("A"), false},
	})
	assertEqual(t, "b", DocToString(doc))
	Delete(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
	})
	assertEqual(t, "", DocToString(doc))
}

func TestInsertLongPath(t *testing.T) {
	doc := NewDisambiguatorNode()
	Insert(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
		Dir{[]byte("A"), false},
	}, 'c')
	assertEqual(t, "c", DocToString(doc))
	nodes := DocToNodes(doc)
	assertEqual(t, 3, len(nodes))
	assertEqual(t, nodes[0].State, UNINITIALIZED)
	assertEqual(t, nodes[1].State, ALIVE)
	assertEqual(t, nodes[1].Atom, byte('c'))
	assertEqual(t, nodes[2].State, UNINITIALIZED)
}

func TestDeleteLongPath(t *testing.T) {
	doc := NewDisambiguatorNode()
	Delete(doc, []Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("B"), false},
		Dir{[]byte("A"), false},
	})
	assertEqual(t, "", DocToString(doc))
	nodes := DocToNodes(doc)
	assertEqual(t, 3, len(nodes))
	assertEqual(t, nodes[0].State, UNINITIALIZED)
	assertEqual(t, nodes[1].State, DEAD)
	assertEqual(t, nodes[2].State, UNINITIALIZED)
}

// idempotency
func TestInsertIdempotency(t *testing.T) {
	doc := NewDisambiguatorNode()
	Insert(doc, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	assertEqual(t, "a", DocToString(doc))
	Insert(doc, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	assertEqual(t, "a", DocToString(doc))
}

func TestDeleteIdempotency(t *testing.T) {
	doc := NewDisambiguatorNode()
	Insert(doc, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	assertEqual(t, "a", DocToString(doc))
	Delete(doc, []Dir{
		Dir{[]byte("A"), false},
	})
	assertEqual(t, "", DocToString(doc))
	Delete(doc, []Dir{
		Dir{[]byte("A"), false},
	})
	assertEqual(t, "", DocToString(doc))
}

//commutative
func TestInsertInsert(t *testing.T) {
	doc1 := NewDisambiguatorNode()
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, 'b')

	doc2 := NewDisambiguatorNode()
	Insert(doc2, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, 'b')
	Insert(doc2, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')

	assertEqual(t, DocToString(doc1), DocToString(doc2))
}

func TestInsertDelete(t *testing.T) {
	doc1 := NewDisambiguatorNode()
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, 'b')
	Delete(doc1, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	})

	doc2 := NewDisambiguatorNode()
	Insert(doc2, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	Delete(doc2, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	})
	Insert(doc2, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, 'b')

	assertEqual(t, DocToString(doc1), DocToString(doc2))
}

func TestDeleteDelete(t *testing.T) {
	doc1 := NewDisambiguatorNode()
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, 'b')

	doc2 := NewDisambiguatorNode()
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
	}, 'a')
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, 'b')

	Delete(doc1, []Dir{
		Dir{[]byte("A"), false},
	})
	Delete(doc1, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	})

	Delete(doc2, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	})
	Delete(doc2, []Dir{
		Dir{[]byte("A"), false},
	})

	assertEqual(t, DocToString(doc1), DocToString(doc2))
}

func TestGenerateDoc(t *testing.T) {
	doc := GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
	}, "abc")
	assertEqual(t, "abc", DocToString(doc))
	assertEqual(t, 2, Height(doc))

	doc = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
	}, "abcde")
	assertEqual(t, "abcde", DocToString(doc))
	assertEqual(t, 3, Height(doc))

	doc = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
	}, "abcdefg")
	assertEqual(t, "abcdefg", DocToString(doc))
	assertEqual(t, 3, Height(doc))

	doc = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, "abcde")
	assertEqual(t, "abcde", DocToString(doc))
	assertEqual(t, 4, Height(doc))
}

func TestMerge(t *testing.T) {
	doc1 := GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
	}, "abc")
	doc2 := GenerateDoc([]Dir{
		Dir{[]byte("B"), false},
	}, "def")
	assertEqual(t, "abc", DocToString(doc1))
	assertEqual(t, "def", DocToString(doc2))
	Merge(doc1, doc2)
	assertEqual(t, "abcdef", DocToString(doc1))

	doc1 = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
	}, "abc")
	doc2 = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, "def")
	assertEqual(t, "abc", DocToString(doc1))
	assertEqual(t, "def", DocToString(doc2))
	Merge(doc1, doc2)
	assertEqual(t, "defabc", DocToString(doc1))

	doc1 = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
	}, "abc")
	doc2 = GenerateDoc([]Dir{
		Dir{[]byte("A"), true},
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, "def")
	assertEqual(t, "abc", DocToString(doc1))
	assertEqual(t, "def", DocToString(doc2))
	Merge(doc1, doc2)
	assertEqual(t, "abdefc", DocToString(doc1))

	doc1 = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
	}, "abc")
	Insert(doc1, []Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), true},
		Dir{[]byte("A"), false},
	}, 'A')
	doc2 = GenerateDoc([]Dir{
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), true},
		Dir{[]byte("A"), false},
		Dir{[]byte("A"), false},
	}, "def")
	assertEqual(t, "aAbc", DocToString(doc1))
	assertEqual(t, "def", DocToString(doc2))
	Merge(doc1, doc2)
	assertEqual(t, "adefAbc", DocToString(doc1))
}
