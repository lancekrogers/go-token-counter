package spm

import "unicode/utf8"

// prefixMatcher is a trie-based longest prefix matcher.
type prefixMatcher struct {
	root *trieNode
}

type trieNode struct {
	children map[rune]*trieNode
	final    bool
}

func newPrefixMatcher(vocab map[string]bool) *prefixMatcher {
	pm := &prefixMatcher{root: newTrieNode()}
	for word := range vocab {
		pm.add(word)
	}
	return pm
}

// findPrefixLen finds the longest prefix of text matching a vocab word.
func (pm *prefixMatcher) findPrefixLen(text string) int {
	node := pm.root
	maxLen := 0

	for i, r := range text {
		child := node.children[r]
		if child == nil {
			return maxLen
		}
		if child.final {
			maxLen = i + utf8.RuneLen(r)
		}
		node = child
	}

	return maxLen
}

func (pm *prefixMatcher) add(word string) {
	node := pm.root

	for _, r := range word {
		child := node.children[r]
		if child == nil {
			child = newTrieNode()
			node.children[r] = child
		}
		node = child
	}

	node.final = true
}

func newTrieNode() *trieNode {
	return &trieNode{
		children: make(map[rune]*trieNode),
		final:    false,
	}
}
