package sfuzz

import (
	"context"
	"fmt"
	"iter"
)

type (
	FuzzFunc func(context.Context) (any, bool)
	Selector func(FuzzKeyword) FuzzFunc
)

func NumFuzzer() FuzzFunc          { return FuzzFromList(numList) }
func UIDFuzzer() FuzzFunc          { return FuzzFromList(uidList) }
func StringFuzzer() FuzzFunc       { return FuzzFromList(strList) }
func StringInPathFuzzer() FuzzFunc { return FuzzFromList(strInPathList) }

func noopFuzzer(context.Context) (any, bool) { return nil, false }
func noopSelector() Selector {
	return func(FuzzKeyword) FuzzFunc {
		return noopFuzzer
	}
}

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

var DefaultSelector = func(k FuzzKeyword) FuzzFunc {
	switch k.Kind {
	case Numeral:
		return NumFuzzer()
	case UniversalID:
		return UIDFuzzer()
	case GenericString:
		switch k.Location {
		case PathKeyword:
			return StringInPathFuzzer()
		default:
			return StringFuzzer()
		}
	}
	return noopFuzzer
}

func WordlistSelector(lists map[Kind][]any) Selector {
	if lists == nil || len(lists) == 0 {
		return noopSelector()
	}
	return func(k FuzzKeyword) FuzzFunc {
		return FuzzFromList(lists[k.Kind])
	}
}

// CounterFuzzer is mostly use for predictable outcome in tests
func CounterFuzzer(count int) FuzzFunc {
	var out []any
	for i := range count {
		out = append(out, fmt.Sprintf("counter_%d", i))
	}
	return FuzzFromList(out)
}
