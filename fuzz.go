package sfuzz

import "iter"

type Generator func(string) iter.Seq[any]

type Selector func(FuzzKeyword) Generator

// StableGenerator is mostly use for predictable outcome in tests
func StableGenerator(count int) Generator {
	return func(s string) iter.Seq[any] {
		return func(yield func(any) bool) {
			for range count {
				if !yield(s) {
					return
				}
			}
		}
	}
}

func NoopGenerator() Generator {
	return func(s string) iter.Seq[any] {
		return func(yield func(any) bool) { return }
	}
}
