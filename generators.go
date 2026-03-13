package sfuzz

import (
	"context"
	"fmt"
	"iter"
	"math"
	"strings"

	"github.com/google/uuid"
)

type Iterator interface {
	Next(context.Context) (any, bool)
}

func NewIterator(g Generator) Iterator {
	next, stop := iter.Pull(g)
	return ListIterator{next: next, stop: stop}
}

type ListIterator struct {
	next func() (any, bool)
	stop func()
}

func (l ListIterator) Next(ctx context.Context) (any, bool) {
	select {
	case <-ctx.Done():
		l.stop()
		return nil, false
	default:
		return l.next()
	}
}

func FromList(list []any) Generator {
	return func(yield func(any) bool) {
		for _, n := range list {
			if !yield(n) {
				return
			}
		}
	}
}

func NumGenerator() Generator { return FromList(numList) }
func UIDGenerator() Generator { return FromList(uidList) }
func StrGenerator() Generator { return FromList(strList) }

var strList = []any{
	"", " ", "  ", "   ",
	".", "..", "...",
}

var numList = []any{
	math.MaxInt64,
	math.MinInt64,
	fmt.Sprintf("%d00", uint64(math.MaxUint64)),
	fmt.Sprintf("%d00", math.MinInt64),
	0, 0.00, -1.00, -1.0,
	1e23, -1e23,
}

var uidList = []any{
	uuid.Nil.String(),
	fmt.Sprintf("%s%s", uuid.New().String(), uuid.New().String()),
	strings.ReplaceAll(uuid.New().String(), "-", "."),
	strings.ReplaceAll(uuid.New().String(), "-", ""),
}
