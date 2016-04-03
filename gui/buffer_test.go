package gui

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

// not very good tests

func TestStringToBuffer(t *testing.T) {
	buf := StringToBuffer("abc abc\ndef def", 10)
	assertEqual(t, 2, NumberOfLines(buf.lines))
	assertEqual(t, "abc abc\ndef def", buf.ToString())

	buf = StringToBuffer("abc abc\ndef def def", 10)
	assertEqual(t, 3, NumberOfLines(buf.lines))
	assertEqual(t, "abc abc\ndef def def", buf.ToString())

	buf = StringToBuffer("abc abc\n\tdef def", 10)
	assertEqual(t, 3, NumberOfLines(buf.lines))
	assertEqual(t, "abc abc\n\tdef def", buf.ToString())
}
