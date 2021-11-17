package util

import (
	"strings"
	"testing"
)

func TestPascalToSnake(t *testing.T) {
	Expect := func(src, dst string) {
		if actual := PascalToSnake(src); actual != dst {
			t.Errorf("PascalToSnake(\"%s\") expects \"%s\" but actual \"%s\"", src, dst, actual)
		}
	}
	Expect("ABCDEF", "a_b_c_d_e_f")
	Expect("HelloWorld", "hello_world")
}

func TestFuncNameToPath(t *testing.T) {
	Expect := func(src, dst string) {
		if actual := FuncNameToPath(src); actual != dst {
			t.Errorf("FuncNameToPath(\"%s\") expects \"%s\" but actual \"%s\"", src, dst, actual)
		}
	}
	Expect("ABCDEF", "/a_b_c_d_e_f")
	Expect("Test_nameis_HelloWorld", "/test/nameis/hello_world")
}

func TestPascalSplit(t *testing.T) {
	Expect := func(src, dst string) {
		if actual := PascalSplit(src); strings.Join(actual, "_") != dst {
			t.Errorf("PascalToSnake(\"%s\") expects \"%s\" but actual \"%s\"", src, dst, actual)
		}
	}
	Expect("ABCDEF", "a_b_c_d_e_f")
	Expect("HelloWorld", "hello_world")
}
