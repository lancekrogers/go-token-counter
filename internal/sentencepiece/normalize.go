package sentencepiece

import "strings"

const whitespaceSeparator = "▁"

func normalize(text string) string {
	return replaceSpacesBySeparator(text)
}

func replaceSpacesBySeparator(text string) string {
	return strings.ReplaceAll(text, " ", whitespaceSeparator)
}

func replaceSeparatorsBySpace(text string) string {
	return strings.ReplaceAll(text, whitespaceSeparator, " ")
}
