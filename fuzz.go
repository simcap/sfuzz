package sfuzz

import (
	"fmt"
	"iter"
)

type Generator = iter.Seq[any]

type Selector func(FuzzKeyword) Generator

// CounterGenerator is mostly use for predictable outcome in tests
func CounterGenerator(count int) Generator {
	return func(yield func(any) bool) {
		for i := range count {
			if !yield(fmt.Sprintf("counter_%d", i)) {
				return
			}
		}
	}
}

func NoopGenerator() Generator {
	return func(yield func(any) bool) { return }
}
