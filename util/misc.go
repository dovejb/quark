package util

import "encoding/json"

func Js(x interface{}) string {
	b, e := json.Marshal(x)
	if e != nil {
		panic(e)
	}
	return string(b)
}
