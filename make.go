package quark

import (
	"reflect"
)

// x must be pointer of slice
func MakeSlice(x interface{}, len int) {
	v := reflect.ValueOf(x)
	v.Elem().Set(reflect.MakeSlice(v.Elem().Type(), len, len))
}

func Resize(x interface{}, len int) {
	v := reflect.ValueOf(x)
	src := v.Elem()
	dst := reflect.MakeSlice(v.Elem().Type(), len, len)
	if !src.IsNil() && src.Len() > 0 {
		reflect.Copy(dst, src)
	}
	v.Elem().Set(dst)
}
