package util

import (
	"fmt"
	"strings"
)

type Trie struct {
	value    int
	subs     map[string]Trie
	valid    bool
	notEmpty bool
}

var (
	NilTrie = Trie{}
)

func NewTrie() Trie {
	return Trie{
		subs:  make(map[string]Trie),
		valid: true,
	}
}

func (t Trie) String() string {
	keys := make([]string, 0, len(t.subs))
	for k := range t.subs {
		keys = append(keys, k)
	}
	return fmt.Sprint(keys)
}

func (t Trie) sub(key string) Trie {
	return t.subs[key]
}

func (t *Trie) set(value int) {
	t.notEmpty = true
	t.value = value
}

func (t Trie) Add(path []string, value int) {
	var parent Trie
	st := t
	for i := range path {
		parent = st
		st = st.sub(path[i])
		if !st.valid {
			st = NewTrie()
			parent.subs[path[i]] = st
		}
	}
	st.set(value)
	parent.subs[path[len(path)-1]] = st
}

func (t Trie) Sub(key string) Trie {
	return t.sub(key)
}

func (t Trie) Valid() bool {
	return t.valid
}

func (t Trie) Value() *int {
	if t.notEmpty {
		k := t.value
		return &k
	}
	return nil
}

func (t Trie) Empty() bool {
	return !t.notEmpty
}

func (t Trie) Find(path []string) (result Trie, found bool) {
	st := t
	for i := range path {
		st = st.Sub(path[i])
		if !st.Valid() {
			return t, false
		}
	}
	return st, true
}

// do returns bool means: CONTINUE or not
func (t Trie) Traverse(do func([]string, Trie) bool) {
	t.traverse(make([]string, 0), do)
}

func (t Trie) traverse(prefix []string, do func([]string, Trie) bool) bool {
	if t.notEmpty {
		if !do(prefix, t) {
			return false
		}
	}
	for k, st := range t.subs {
		if !st.traverse(append(prefix, k), do) {
			return false
		}
	}
	return true
}

func (t Trie) Map() map[string]int {
	m := make(map[string]int)
	t.Traverse(func(path []string, trie Trie) bool {
		m[strings.Join(path, "/")] = trie.value
		return true
	})
	return m
}

func (t Trie) Dump() {
	m := t.Map()
	for k, v := range m {
		fmt.Println(k, v)
	}
}
