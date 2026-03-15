package sfuzz_test

import (
	"slices"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestGenerator(t *testing.T) {
	t.Run("stable", func(t *testing.T) {
		var actual []any
		count := 5
		fuzzer := sfuzz.CounterFuzzer(count)
		for range count {
			v, _ := fuzzer.Next(t.Context())
			actual = append(actual, v)
		}
		expected := []any{"counter_0", "counter_1", "counter_2", "counter_3", "counter_4"}
		if !slices.Equal(actual, expected) {
			t.Fatalf("\n got: %v\nwant: %v\n", actual, expected)
		}
	})
}

func TestIterator(t *testing.T) {
	var actual []any
	gen := sfuzz.CounterFuzzer(5)
	for {
		v, more := gen.Next(t.Context())
		if !more {
			break
		}
		actual = append(actual, v)
	}
	expected := []any{"counter_0", "counter_1", "counter_2", "counter_3", "counter_4"}
	if !slices.Equal(actual, expected) {
		t.Fatalf("\n got: %v\nwant: %v\n", actual, expected)
	}
}
