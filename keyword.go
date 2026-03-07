package sfuzz

import (
	"fmt"
	"slices"
	"strings"
)

var (
	FuzzPrefix     = "FUZZ"
	TypeSuffixSize = 3

	GenericString Kind = "STR"
	Numeral       Kind = "NUM"
	UniversalID   Kind = "UID"
	Date          Kind = "DTE"
	Time          Kind = "TME"
	AllKinds           = []Kind{GenericString, Numeral, UniversalID, Date, Time}
)

type (
	Kind     string
	Location int
)

const (
	PathKeyword Location = iota
	QueryKeyword
	BodyKeyword
)

type FuzzKeyword struct {
	Start, End int
	Location   Location
	Kind       Kind
	Example    string
}

func ParseKeywords(input string) ([]FuzzKeyword, error) {
	var (
		out    []FuzzKeyword
		offset int
	)
	for {
		index := strings.Index(input[offset:], FuzzPrefix)
		if index < 0 {
			return out, nil
		}

		keyword, err := parseKeyword(input, index+offset)
		if err != nil {
			return out, err
		}
		out = append(out, keyword)
		offset = keyword.End
	}
}

func parseKeyword(s string, index int) (FuzzKeyword, error) {
	for i := index; i <= len(s)-TypeSuffixSize; i++ {
		kind := Kind(s[i : i+TypeSuffixSize])
		if slices.Contains(AllKinds, kind) {
			return buildKeyword(kind, s, index, i), nil
		}
	}
	return FuzzKeyword{}, fmt.Errorf("no keyword type found: %s", s)
}

func buildKeyword(kind Kind, s string, index, length int) FuzzKeyword {
	start, end := index, length+TypeSuffixSize
	example := s[start+len(FuzzPrefix) : length]
	return FuzzKeyword{Kind: kind, Start: start, End: end, Example: example}
}
