package quark

import (
	"reflect"
	"testing"
	"unicode"
)

type MyService struct {
	Console
	MyAnonymous
}

type MyAnonymous struct {
}

func (ma MyAnonymous) Show() {
}

func (x MyService) POST_HelloWorld_Is_Nice(req struct {
	Score   float64
	Message string
}) (rsp struct {
	Data string
}, e error) {
	return
}

func TestMethods(t *testing.T) {
	typ := reflect.TypeOf(MyService{}.MyAnonymous)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		t.Log(m.Index)
		t.Log(m.Name)
		t.Log(m.PkgPath)
		t.Log(m.Type)
		t.Log(m.Type.In(0))
	}
}

func TestPrivateType(t *testing.T) {
	type private struct{}
	typ := reflect.TypeOf(private{})
	type Public struct{}
	typ2 := reflect.TypeOf(Public{})
	t.Log(typ.Name(), typ.PkgPath())
	t.Log(typ2.Name(), typ2.PkgPath())
}

func TestPathElementChoices(t *testing.T) {
	t.Log(". is symbol?: ", unicode.IsSymbol('.'))
	t.Log(PathElementChoices("123.5"))
	//t.Log(PathElementChoices("中国abc"))
}
