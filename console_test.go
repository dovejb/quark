package quark

import (
	"reflect"
	"testing"
)

func TestIsConsoleMethod(t *testing.T) {
	s := struct {
		Console
	}{}
	typ := reflect.TypeOf(s)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if !isConsoleMethod(method) {
			t.Errorf("Method %v should be console method but not checked out", method)
		}
		t.Logf("Method %v is Console Method", method)
	}
}
