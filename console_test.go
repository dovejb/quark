package quark

import (
	"encoding/json"
	"fmt"
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

func TestMarshalError(t *testing.T) {
	e := fmt.Errorf("TestError %d %s", 123, "hello")
	j, err := json.Marshal(e)
	if err != nil {
		t.Fatal(e)
	}
	t.Log(string(j))
}
