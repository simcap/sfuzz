package sfuzz_test

import (
	"slices"
	"testing"

	"github.com/simcap/sfuzz"
)

func TestCounterFuzzer(t *testing.T) {
	var actual []any
	fuzz := sfuzz.CounterFuzzer(5)
	for {
		v, more := fuzz(t.Context())
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
