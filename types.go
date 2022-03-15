package quark

import (
	"reflect"
)

type String string
type Int int
type Number float64

func (s String) V() string {
	return string(s)
}

func (i Int) V() int {
	return int(i)
}

func (n Number) V() float64 {
	return float64(n)
}

var (
	StringType = reflect.TypeOf(String(""))
	IntType    = reflect.TypeOf(Int(0))
	NumberType = reflect.TypeOf(Number(0))
)

func IsUrlType(t reflect.Type) bool {
	return t == StringType || t == IntType || t == NumberType
}
