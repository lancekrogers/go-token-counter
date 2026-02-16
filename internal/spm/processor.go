package spm

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/lancekrogers/go-token-counter/internal/spm/spmmodel"
	"google.golang.org/protobuf/proto"
)

// Tokenizer implements SentencePiece BPE tokenization.
type Tokenizer struct {
	mdl *spmmodel.ModelProto

	pieces   map[string]int
	reserved map[string]int

	unknownID          int
	userDefinedMatcher *prefixMatcher
	byte2Token         map[byte]Token
	idToByte           map[int]byte
	maxPieceLength     int
}

// NewTokenizerFromPath creates a new Tokenizer from a .model file path.
func NewTokenizerFromPath(protoFile string) (*Tokenizer, error) {
	f, err := os.Open(protoFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read %q: %v", protoFile, err)
	}
	defer f.Close()
	return NewTokenizer(f)
}

// NewTokenizer creates a new Tokenizer from a reader with protobuf data.
func NewTokenizer(protoReader io.Reader) (*Tokenizer, error) {
	b, err := io.ReadAll(protoReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read protobuf data: %v", err)
	}

	var mp spmmodel.ModelProto
	err = proto.Unmarshal(b, &mp)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal protobuf: %v", err)
	}

	tspec := mp.GetTrainerSpec()
	if tspec.GetModelType() != spmmodel.TrainerSpec_BPE {
		return nil, fmt.Errorf("model type %s not supported", tspec.GetModelType())
	}

	nspec := mp.GetNormalizerSpec()
	if *nspec.AddDummyPrefix || *nspec.RemoveExtraWhitespaces {
		return nil, fmt.Errorf("normalizer spec options not supported: %s", nspec)
	}

	userDefined := make(map[string]bool)
	pieces := make(map[string]int)
	reserved := make(map[string]int)
	byte2Token := make(map[byte]Token)
	idToByte := make(map[int]byte)
	unkID := -1
	maxPieceLength := 0

	for i, piece := range mp.GetPieces() {
		isNormalPiece := (piece.GetType() == spmmodel.ModelProto_SentencePiece_NORMAL ||
			piece.GetType() == spmmodel.ModelProto_SentencePiece_USER_DEFINED ||
			piece.GetType() == spmmodel.ModelProto_SentencePiece_UNUSED)

		if isNormalPiece {
			pieces[piece.GetPiece()] = i
			maxPieceLength = max(maxPieceLength, len(piece.GetPiece()))
		} else {
			reserved[piece.GetPiece()] = i
		}

		if piece.GetType() == spmmodel.ModelProto_SentencePiece_USER_DEFINED {
			userDefined[piece.GetPiece()] = true
		} else if piece.GetType() == spmmodel.ModelProto_SentencePiece_UNKNOWN {
			if unkID > 0 {
				return nil, fmt.Errorf("unk redefined")
			}
			unkID = i
		} else if piece.GetType() == spmmodel.ModelProto_SentencePiece_BYTE {
			if !tspec.GetByteFallback() {
				return nil, fmt.Errorf("byte piece %q found but byte_fallback=false", piece.GetPiece())
			}
			bv := convertHexValue(piece.GetPiece())
			if bv >= 0 && bv < 256 {
				byte2Token[byte(bv)] = Token{ID: i, Text: piece.GetPiece()}
				idToByte[i] = byte(bv)
			}
		}
	}

	if unkID < 0 {
		return nil, fmt.Errorf("unk symbol is not defined")
	}

	if tspec.GetByteFallback() {
		for i := range 256 {
			if _, found := byte2Token[byte(i)]; !found {
				return nil, fmt.Errorf("byte value 0x%02X not found", i)
			}
		}
	}

	return &Tokenizer{
		mdl:                &mp,
		userDefinedMatcher: newPrefixMatcher(userDefined),
		byte2Token:         byte2Token,
		idToByte:           idToByte,
		unknownID:          unkID,
		pieces:             pieces,
		reserved:           reserved,
		maxPieceLength:     maxPieceLength,
	}, nil
}

// Encode tokenizes the input text and returns a list of Tokens.
func (tok *Tokenizer) Encode(text string) []Token {
	text = normalize(text)

	type symListElem struct {
		prev, next int
		noMerge    bool
		symbol     string
	}
	symList := make([]symListElem, 0, len(text))

	for {
		slen, found := tok.symbolMatch(text)
		sym := symListElem{
			noMerge: found,
			symbol:  text[:slen],
			prev:    len(symList) - 1,
			next:    len(symList) + 1,
		}
		symList = append(symList, sym)
		text = text[slen:]
		if len(text) == 0 {
			break
		}
	}

	if len(symList) == 0 {
		return nil
	}
	symList[len(symList)-1].next = -1
	nTokens := len(symList)

	type mergeCandidate struct {
		left, right int
		length      int
		score       float32
	}

	mergeQueue := newPriorityQueue(len(symList), func(a, b mergeCandidate) int {
		if a.score > b.score || (a.score == b.score && a.left < b.left) {
			return 1
		}
		return -1
	})

	buf := make([]byte, tok.maxPieceLength)
	findMerged := func(x, y symListElem) (string, int, bool) {
		combinedLen := len(x.symbol) + len(y.symbol)
		if combinedLen > cap(buf) {
			return "", 0, false
		}
		buf = buf[:combinedLen]
		copy(buf, x.symbol)
		copy(buf[len(x.symbol):], y.symbol)
		if id, found := tok.pieces[string(buf)]; found {
			return tok.mdl.GetPieces()[id].GetPiece(), id, true
		}
		return "", 0, false
	}

	suggestNewMergePair := func(left, right int) {
		if left == -1 || right == -1 || symList[left].noMerge || symList[right].noMerge {
			return
		}
		if mergedSymbol, id, ok := findMerged(symList[left], symList[right]); ok {
			mergeQueue.Insert(mergeCandidate{
				left:   left,
				right:  right,
				length: len(mergedSymbol),
				score:  tok.mdl.GetPieces()[id].GetScore(),
			})
		}
	}

	for i := 1; i < len(symList); i++ {
		suggestNewMergePair(i-1, i)
	}

	candidateIsDead := func(candidate mergeCandidate) bool {
		leftSymbol := symList[candidate.left].symbol
		rightSymbol := symList[candidate.right].symbol
		return leftSymbol == "" || rightSymbol == "" || len(leftSymbol)+len(rightSymbol) != candidate.length
	}

	mergeQueueDead := 0
	for mergeQueue.Len() > 0 {
		candidate := mergeQueue.PopMax()
		leftSymbol := symList[candidate.left]
		rightSymbol := symList[candidate.right]

		if candidateIsDead(candidate) {
			mergeQueueDead--
			continue
		}

		if mergeQueueDead*3 > mergeQueue.Len() {
			mergeQueue.RemoveFunc(candidateIsDead)
			mergeQueueDead = 0
		}

		mergedSymbol, _, ok := findMerged(leftSymbol, rightSymbol)
		if !ok {
			panic("failed to merge symbols")
		}
		symList[candidate.left].symbol = mergedSymbol
		nTokens--

		symList[candidate.left].next = rightSymbol.next
		if rightSymbol.next >= 0 {
			symList[rightSymbol.next].prev = candidate.left
		}

		symList[candidate.right].symbol = ""
		mergeQueueDead++

		suggestNewMergePair(leftSymbol.prev, candidate.left)
		suggestNewMergePair(candidate.left, rightSymbol.next)
	}

	tokens := make([]Token, 0, nTokens)
	for i := 0; i >= 0; i = symList[i].next {
		symbol := symList[i].symbol
		id := tok.symbolToID(symbol)

		if id == tok.unknownID && tok.mdl.GetTrainerSpec().GetByteFallback() {
			for j := range len(symbol) {
				tokens = append(tokens, tok.byte2Token[symbol[j]])
			}
		} else {
			tokens = append(tokens, Token{ID: id, Text: symbol})
		}
	}

	return tokens
}

func (tok *Tokenizer) symbolMatch(text string) (int, bool) {
	prefixLen := tok.userDefinedMatcher.findPrefixLen(text)
	if prefixLen > 0 {
		return prefixLen, true
	}
	_, rlen := utf8.DecodeRuneInString(text)
	return rlen, false
}

func (tok *Tokenizer) symbolToID(symbol string) int {
	if id, found := tok.reserved[symbol]; found {
		return id
	}
	if id, found := tok.pieces[symbol]; found {
		return id
	}
	return tok.unknownID
}

func convertHexValue(bv string) int {
	bv = strings.TrimPrefix(bv, "<0x")
	bv = strings.TrimSuffix(bv, ">")
	n, err := strconv.ParseInt(bv, 16, 32)
	if err != nil {
		return -1
	}
	return int(n)
}

// Decode translates a list of token IDs back into the string they represent.
func (tok *Tokenizer) Decode(ids []int) string {
	var sb strings.Builder

	for i := 0; i < len(ids); {
		nextNonByte := i
		for nextNonByte < len(ids) && tok.isByteID(ids[nextNonByte]) {
			nextNonByte++
		}
		numBytes := nextNonByte - i

		if numBytes > 0 {
			buf := make([]byte, 0, numBytes)
			for bi := i; bi < nextNonByte; bi++ {
				buf = append(buf, tok.idToByte[ids[bi]])
			}
			for len(buf) > 0 {
				r, size := utf8.DecodeRune(buf)
				sb.WriteRune(r)
				buf = buf[size:]
			}
		}

		if nextNonByte >= len(ids) {
			break
		}
		id := ids[nextNonByte]
		if tok.isControlID(id) {
			// Don't emit control IDs
		} else if id == tok.unknownID {
			sb.WriteString(tok.mdl.GetTrainerSpec().GetUnkSurface())
		} else {
			piece := tok.mdl.GetPieces()[id].GetPiece()
			sb.WriteString(replaceSeparatorsBySpace(piece))
		}
		i = nextNonByte + 1
	}

	return sb.String()
}

// DecodeTokens is a convenience wrapper around Decode.
func (tok *Tokenizer) DecodeTokens(tokens []Token) string {
	ids := make([]int, len(tokens))
	for i, t := range tokens {
		ids[i] = t.ID
	}
	return tok.Decode(ids)
}

func (tok *Tokenizer) isByteID(id int) bool {
	return tok.mdl.GetPieces()[id].GetType() == spmmodel.ModelProto_SentencePiece_BYTE
}

func (tok *Tokenizer) isControlID(id int) bool {
	return tok.mdl.GetPieces()[id].GetType() == spmmodel.ModelProto_SentencePiece_CONTROL
}
