package bpe

import (
	"embed"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

//go:embed vocabdata/*.tiktoken
var vocabFS embed.FS

// loadEmbeddedVocab loads BPE ranks from an embedded vocab file.
func loadEmbeddedVocab(name string) (map[string]int, error) {
	data, err := vocabFS.ReadFile("vocabdata/" + name + ".tiktoken")
	if err != nil {
		return nil, fmt.Errorf("embedded vocab %q not found: %w", name, err)
	}
	return parseBPERanks(data)
}

// parseBPERanks parses tiktoken-format BPE data into a rank map.
func parseBPERanks(data []byte) (map[string]int, error) {
	ranks := make(map[string]int)
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}
		token, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, fmt.Errorf("decoding token: %w", err)
		}
		rank, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("parsing rank: %w", err)
		}
		ranks[string(token)] = rank
	}
	return ranks, nil
}
