package sfuzz

import "iter"

func Fuzz(value string, kinds ...Kind) {

}

type Generator func(s string) iter.Seq[string]

// StableGenerator is mostly use for predictable outcome in tests
func StableGenerator(count int) Generator {
	return func(s string) iter.Seq[string] {
		return func(yield func(string) bool) {
			for range count {
				if !yield(s) {
					return
				}
			}
		}
	}
}
