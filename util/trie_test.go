package util

import "testing"

func TestTrie(t *testing.T) {
	trie := NewTrie()
	trie.Add([]string{"hello", "world", "mary", "tony"}, 1)
	trie.Add([]string{"hello", "world", "mary", "bob"}, 2)
	t.Log(trie.Map())
	trie.Dump()

	for k := range trie.subs {
		t.Log(k)
	}
	t.Log(trie.String())
}
