package quark

import (
	"reflect"
)

type String string
type Int int
type Number float64

var (
	StringType = reflect.TypeOf(String(""))
	IntType    = reflect.TypeOf(Int(0))
	NumberType = reflect.TypeOf(Number(0))
)

func IsUrlType(t reflect.Type) bool {
	return t == StringType || t == IntType || t == NumberType
}
