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

// basic op
func TestInsert(t *testing.T) {
}

func TestDelete(t *testing.T) {
}

func TestInsertLongPath(t *testing.T) {
}

func TestDeleteLongPath(t *testing.T) {
}

// idempotency
func TestInsertIdempotency(t *testing.T) {
}

func TestDeleteIdempotency(t *testing.T) {
}

//commutative
func TestInsertInsert(t *testing.T) {
}

func TestInsertDelete(t *testing.T) {
}

func TestDeleteDelete(t *testing.T) {
}

func TestGenerateDoc(t *testing.T) {
}

func TestMerge(t *testing.T) {
}
