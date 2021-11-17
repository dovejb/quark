package util

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

func PascalSplit(s string) (ss []string) {
	start := 0
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i != start {
				ss = append(ss, strings.ToLower(s[start:i]))
				start = i
			}
		}
	}
	if start < len(s) {
		ss = append(ss, strings.ToLower(s[start:]))
	}
	return ss
}

func PascalToSnake(s string) string {
	b := strings.Builder{}
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func CamelToSnake(s string) string {
	return PascalToSnake(s)
}

func FuncNameToPath(s string) string {
	ss := strings.Split(s, "_")
	for i := range ss {
		ss[i] = PascalToSnake(ss[i])
	}
	return "/" + strings.Join(ss, "/")
}

type PathVar struct {
	Pos int    //starts from 0
	Var string //typeChar + '.' + varname
}

/*
	path element,
 		upper leading - path component
		lower leading - path variable
	{s} - string
	{i} - integer
	{f} - number
*/
func FuncNameToPathWithVars(s string, varsType []reflect.Kind) (pathFormat string, vars []PathVar, e error) {
	ss := strings.Split(s, "_")
	vti := 0
	for i := range ss {
		rs := []rune(ss[i])
		if unicode.IsUpper(rs[0]) {
			pathFormat += "/" + PascalToSnake(ss[i])
		} else {
			if vti >= len(varsType) {
				e = fmt.Errorf("parsing path variables: name variable num > in parameter num")
				return
			}
			var typeChar string
			switch varsType[vti] {
			case reflect.String:
				typeChar = "s"
			case reflect.Int:
				typeChar = "i"
			case reflect.Float64:
				typeChar = "f"
			default:
				e = fmt.Errorf("parsing path variables: invalid varsType element: %v", varsType[vti])
				return
			}
			pathFormat += "/{" + typeChar + "}"
			vars = append(vars, PathVar{i, typeChar + "." + CamelToSnake(ss[i])})
			vti += 1
		}
	}
	if len(varsType) != len(vars) {
		e = fmt.Errorf("parsing path variables: name variable num != in parameter num")
		return
	}
	return
}
