package sfuzz

import (
	"fmt"
	"iter"
	"math"
	"strings"

	"github.com/google/uuid"
)

func FromList(list []any) Generator {
	return func(string) iter.Seq[any] {
		return func(yield func(any) bool) {
			for _, n := range list {
				if !yield(n) {
					return
				}
			}
		}
	}
}

func NumGenerator(string) Generator { return FromList(numList) }
func UIDGenerator(string) Generator { return FromList(uidList) }
func StrGenerator(string) Generator { return FromList(strList) }

var strList = []any{
	"", " ", "  ", "   ",
	".", "..", "...",
}

var numList = []any{
	math.MaxInt64,
	math.MinInt64,
	0, 0.00, -1.00, -1.0,
	1e23, -1e23,
}

var uidList = []any{
	uuid.Nil.String(),
	fmt.Sprintf("%s%s", uuid.New().String(), uuid.New().String()),
	strings.ReplaceAll(uuid.New().String(), "-", "."),
	strings.ReplaceAll(uuid.New().String(), "-", ""),
}
