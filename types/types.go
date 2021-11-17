package types

import (
	"math/big"
	"reflect"
)

type URLString string
type URLInt int
type URLNumber big.Float

var (
	URLStringType = reflect.TypeOf(URLString(""))
	URLIntType    = reflect.TypeOf(URLInt(0))
	URLNumberType = reflect.TypeOf(URLNumber{})
)

func IsUrlType(t reflect.Type) bool {
	return t == URLStringType || t == URLIntType || t == URLNumberType
}
