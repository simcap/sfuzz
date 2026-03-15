package sfuzz

import (
	"context"
	"fmt"
	"iter"
	"math"
	"strings"

	"github.com/google/uuid"
)

type Fuzzer interface {
	Next(context.Context) (any, bool)
	Stop()
}

func NewFuzzerFromList(list []any) Fuzzer {
	seq := fromList(list)
	next, stop := iter.Pull(seq)
	return listIterator{next: next, stop: stop}
}

type listIterator struct {
	next func() (any, bool)
	stop func()
}

func (l listIterator) Next(ctx context.Context) (any, bool) {
	select {
	case <-ctx.Done():
		l.stop()
		return nil, false
	default:
		return l.next()
	}
}

func (l listIterator) Stop() { l.stop() }

func NumGenerator() Fuzzer { return NewFuzzerFromList(numList) }
func UIDGenerator() Fuzzer { return NewFuzzerFromList(uidList) }
func StrGenerator() Fuzzer { return NewFuzzerFromList(strList) }

func fromList(list []any) iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, n := range list {
			if !yield(n) {
				return
			}
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

type Selector func(FuzzKeyword) Fuzzer

// CounterFuzzer is mostly use for predictable outcome in tests
func CounterFuzzer(count int) Fuzzer {
	var out []any
	for i := range count {
		out = append(out, fmt.Sprintf("counter_%d", i))
	}
	return NewFuzzerFromList(out)
}

func NoopFuzzer() Fuzzer {
	return noopFuzzer{}
}

type noopFuzzer struct{}

func (n noopFuzzer) Next(context.Context) (any, bool) { return nil, false }
func (n noopFuzzer) Stop()                            {}
