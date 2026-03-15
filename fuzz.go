package sfuzz

import (
	"context"
	"fmt"
	"iter"
	"math"
	"strings"

	"github.com/google/uuid"
)

type (
	FuzzFunc func(context.Context) (any, bool)
	Selector func(FuzzKeyword) FuzzFunc
)

func NumFuzzer() FuzzFunc    { return FuzzFromList(numList) }
func UIDFuzzer() FuzzFunc    { return FuzzFromList(uidList) }
func StringFuzzer() FuzzFunc { return FuzzFromList(strList) }

// CounterFuzzer is mostly use for predictable outcome in tests
func CounterFuzzer(count int) FuzzFunc {
	var out []any
	for i := range count {
		out = append(out, fmt.Sprintf("counter_%d", i))
	}
	return FuzzFromList(out)
}

func noopFuzzer(context.Context) (any, bool) { return nil, false }

func FuzzFromList(list []any) FuzzFunc {
	seq := func(yield func(any) bool) {
		for _, n := range list {
			if !yield(n) {
				return
			}
		}
	}

	next, stop := iter.Pull(seq)
	return func(ctx context.Context) (any, bool) {
		select {
		case <-ctx.Done():
			stop()
			return nil, false
		default:
			return next()
		}
	}
}

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
