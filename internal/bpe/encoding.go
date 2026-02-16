package bpe

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
)

// Special token constants.
const (
	EndOfText   = "<|endoftext|>"
	FIMPrefix   = "<|fim_prefix|>"
	FIMMiddle   = "<|fim_middle|>"
	FIMSuffix   = "<|fim_suffix|>"
	EndOfPrompt = "<|endofprompt|>"
)

// Encoding names.
const (
	EncodingO200kBase  = "o200k_base"
	EncodingCL100kBase = "cl100k_base"
	EncodingP50kBase   = "p50k_base"
	EncodingP50kEdit   = "p50k_edit"
	EncodingR50kBase   = "r50k_base"
)

// Encoding holds the definition for a BPE encoding.
type Encoding struct {
	Name           string
	PatStr         string
	MergeableRanks map[string]int
	SpecialTokens  map[string]int
	ExplicitNVocab int
}

var (
	encodingMap = make(map[string]*Encoding)
	mu          sync.RWMutex
)

func getEncoding(encodingName string) (*Encoding, error) {
	mu.RLock()
	enc, ok := encodingMap[encodingName]
	mu.RUnlock()
	if ok {
		return enc, nil
	}

	mu.Lock()
	defer mu.Unlock()

	// Double-check after acquiring write lock.
	if enc, ok := encodingMap[encodingName]; ok {
		return enc, nil
	}

	enc, err := initEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	encodingMap[encodingName] = enc
	return enc, nil
}

func initEncoding(encodingName string) (*Encoding, error) {
	switch encodingName {
	case EncodingO200kBase:
		return o200kBase()
	case EncodingCL100kBase:
		return cl100kBase()
	case EncodingP50kBase:
		return p50kBase()
	case EncodingR50kBase:
		return r50kBase()
	case EncodingP50kEdit:
		return p50kEdit()
	default:
		return nil, errors.New("unknown encoding: " + encodingName)
	}
}

func o200kBase() (*Encoding, error) {
	ranks, err := loadEmbeddedVocab("o200k_base")
	if err != nil {
		return nil, err
	}
	pats := []string{
		`[^\r\n\p{L}\p{N}]?[\p{Lu}\p{Lt}\p{Lm}\p{Lo}\p{M}]*[\p{Ll}\p{Lm}\p{Lo}\p{M}]+(?i:'s|'t|'re|'ve|'m|'ll|'d)?`,
		`[^\r\n\p{L}\p{N}]?[\p{Lu}\p{Lt}\p{Lm}\p{Lo}\p{M}]+[\p{Ll}\p{Lm}\p{Lo}\p{M}]*(?i:'s|'t|'re|'ve|'m|'ll|'d)?`,
		`\p{N}{1,3}`,
		` ?[^\s\p{L}\p{N}]+[\r\n/]*`,
		`\s*[\r\n]+`,
		`\s+(?!\S)`,
		`\s+`,
	}
	return &Encoding{
		Name:           EncodingO200kBase,
		PatStr:         strings.Join(pats, "|"),
		MergeableRanks: ranks,
		SpecialTokens:  map[string]int{EndOfText: 199999, EndOfPrompt: 200018},
	}, nil
}

func cl100kBase() (*Encoding, error) {
	ranks, err := loadEmbeddedVocab("cl100k_base")
	if err != nil {
		return nil, err
	}
	return &Encoding{
		Name:           EncodingCL100kBase,
		PatStr:         `(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`,
		MergeableRanks: ranks,
		SpecialTokens: map[string]int{
			EndOfText: 100257, FIMPrefix: 100258,
			FIMMiddle: 100259, FIMSuffix: 100260,
			EndOfPrompt: 100276,
		},
	}, nil
}

func p50kBase() (*Encoding, error) {
	ranks, err := loadEmbeddedVocab("p50k_base")
	if err != nil {
		return nil, err
	}
	return &Encoding{
		Name:           EncodingP50kBase,
		PatStr:         `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`,
		MergeableRanks: ranks,
		SpecialTokens:  map[string]int{EndOfText: 50256},
		ExplicitNVocab: 50281,
	}, nil
}

func p50kEdit() (*Encoding, error) {
	ranks, err := loadEmbeddedVocab("p50k_base")
	if err != nil {
		return nil, err
	}
	return &Encoding{
		Name:           EncodingP50kEdit,
		PatStr:         `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`,
		MergeableRanks: ranks,
		SpecialTokens:  map[string]int{EndOfText: 50256, FIMPrefix: 50281, FIMMiddle: 50282, FIMSuffix: 50283},
	}, nil
}

func r50kBase() (*Encoding, error) {
	ranks, err := loadEmbeddedVocab("r50k_base")
	if err != nil {
		return nil, err
	}
	return &Encoding{
		Name:           EncodingR50kBase,
		PatStr:         `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`,
		MergeableRanks: ranks,
		SpecialTokens:  map[string]int{EndOfText: 50256},
		ExplicitNVocab: 50257,
	}, nil
}

// GetEncoding returns a Tiktoken for the named encoding.
func GetEncoding(encodingName string) (*Tiktoken, error) {
	enc, err := getEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	pbe, err := NewCoreBPE(enc.MergeableRanks, enc.SpecialTokens, enc.PatStr)
	if err != nil {
		return nil, err
	}
	specialTokensSet := make(map[string]any, len(enc.SpecialTokens))
	for k := range enc.SpecialTokens {
		specialTokensSet[k] = true
	}
	return newTiktoken(pbe, enc, specialTokensSet), nil
}

// Tiktoken is the main tokenizer that wraps CoreBPE.
type Tiktoken struct {
	bpe              *CoreBPE
	pbeEncoding      *Encoding
	specialTokensSet map[string]any
}

func newTiktoken(bpe *CoreBPE, encoding *Encoding, specialTokensSet map[string]any) *Tiktoken {
	return &Tiktoken{
		bpe:              bpe,
		pbeEncoding:      encoding,
		specialTokensSet: specialTokensSet,
	}
}

// Encode tokenizes text with optional special token handling.
func (t *Tiktoken) Encode(text string, allowedSpecial []string, disallowedSpecial []string) []int {
	var allowedSpecialSet map[string]any
	if len(allowedSpecial) == 0 {
		allowedSpecialSet = map[string]any{}
	} else if len(allowedSpecial) == 1 && allowedSpecial[0] == "all" {
		allowedSpecialSet = t.specialTokensSet
	} else {
		allowedSpecialSet = make(map[string]any, len(allowedSpecial))
		for _, v := range allowedSpecial {
			allowedSpecialSet[v] = nil
		}
	}

	disallowedSpecialSet := make(map[string]any, len(disallowedSpecial))
	for _, v := range disallowedSpecial {
		disallowedSpecialSet[v] = nil
	}
	if len(disallowedSpecial) == 1 && disallowedSpecial[0] == "all" {
		disallowedSpecialSet = difference(t.specialTokensSet, allowedSpecialSet)
	}

	if len(disallowedSpecialSet) > 0 {
		specialRegex := t.specialTokenRegex(disallowedSpecialSet)
		m := findRegex2StringMatch(text, specialRegex)
		if m != "" {
			panic(fmt.Sprintf("text contains disallowed special token %s", m))
		}
	}

	tokens, _ := t.bpe.encodeNative(text, allowedSpecialSet)
	return tokens
}

// EncodeOrdinary tokenizes text without special token handling.
func (t *Tiktoken) EncodeOrdinary(text string) []int {
	return t.bpe.encodeOrdinaryNative(text)
}

// Decode converts token IDs back to text.
func (t *Tiktoken) Decode(tokens []int) string {
	return string(t.bpe.decodeNative(tokens))
}

func (t *Tiktoken) specialTokenRegex(disallowedSpecialSet map[string]any) *regexp2.Regexp {
	strs := make([]string, 0, len(disallowedSpecialSet))
	for k := range disallowedSpecialSet {
		strs = append(strs, regexp.QuoteMeta(k))
	}
	return regexp2.MustCompile(strings.Join(strs, "|"), regexp2.None)
}

func findRegex2StringMatch(text string, reg *regexp2.Regexp) string {
	m, _ := reg.FindStringMatch(text)
	if m == nil {
		return ""
	}
	return m.String()
}

func difference(setA, setB map[string]any) map[string]any {
	result := make(map[string]any)
	for k := range setA {
		if _, ok := setB[k]; !ok {
			result[k] = true
		}
	}
	return result
}
