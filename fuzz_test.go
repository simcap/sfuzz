package sfuzz_test

import (
	"slices"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestGenerator(t *testing.T) {
	t.Run("stable", func(t *testing.T) {
		var actual []any
		gen := sfuzz.StableGenerator(5)
		for v := range gen("any") {
			actual = append(actual, v)
		}
		expected := []any{"any", "any", "any", "any", "any"}
		if !slices.Equal(actual, expected) {
			t.Fatalf("\n got: %v\nwant: %v\n", actual, expected)
		}
	})
}
